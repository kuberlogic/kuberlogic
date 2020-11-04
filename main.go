package main

import (
	"errors"
	"flag"
	"fmt"
	mysql "github.com/presslabs/mysql-operator/pkg/apis"
	redis "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"gitlab.com/cloudmanaged/operator/monitoring"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"gitlab.com/cloudmanaged/operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(cloudlinuxv1.AddToScheme(scheme))
	utilruntime.Must(postgres.AddToScheme(scheme))
	utilruntime.Must(mysql.AddToScheme(scheme))
	utilruntime.Must(redis.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

var envRequired = []string{
	"IMAGE_REPO",
	"IMAGE_PULL_SECRET",
}

func checkEnv() error {
	for _, envVariable := range envRequired {
		if value := os.Getenv(envVariable); value == "" {
			return errors.New(fmt.Sprintf(
				"required env variable is undefined: %s", envVariable,
			))
		}
	}
	return nil
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	if err := checkEnv(); err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "2f195a6b.cloudlinux.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// create controller for CloudManaged resource
	if err = (&controllers.CloudManagedReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("CloudManaged"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CloudManaged")
		os.Exit(1)
	}
	// init CloudManaged collector
	cloudManagedCollector := monitoring.CloudManagedCollector{}
	metrics.Registry.MustRegister(cloudManagedCollector)

	// create controller for CloudManagedBackup resource
	if err = (&controllers.CloudManagedBackupReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controller-backup").WithName("CloudManagedBackup"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create backup controller",
			"controller-backup", "CloudManagedBackup")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
