package cmd

import (
	"errors"
	"flag"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/controllers"
	"github.com/kuberlogic/operator/modules/operator/logging"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	"github.com/kuberlogic/operator/modules/operator/util"
	mysql "github.com/presslabs/mysql-operator/pkg/apis"
	redis "github.com/spotahome/redis-operator/api/redisfailover/v1"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"time"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogicv1.AddToScheme(scheme))
	utilruntime.Must(postgres.AddToScheme(scheme))
	utilruntime.Must(mysql.AddToScheme(scheme))
	utilruntime.Must(redis.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

var envRequired = []string{
	util.EnvImgRepo,
	util.EnvImgPullSecret,
}

func checkEnv(log logr.Logger) error {
	for _, envVariable := range envRequired {
		if value := os.Getenv(envVariable); value == "" {
			return errors.New(fmt.Sprintf(
				"required env variable is undefined: %s", envVariable,
			))
		} else {
			log.Info(fmt.Sprintf("%s: %s", envVariable, value))
		}
	}
	return nil
}

func Main(args []string) {

	// init sentry

	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	logger, err := logging.CreateLogger()
	if err != nil {
		setupLog.Error(err, "unable to create logger")
		os.Exit(1)
	}
	ctrl.SetLogger(logger)

	// init sentry
	if dsn := os.Getenv("SENTRY_DSN"); dsn != "" {
		err = sentry.Init(sentry.ClientOptions{
			Dsn: dsn,
		})
		if err != nil {
			logger.Error(err, "unable to create sentry logger")
			os.Exit(1)
		}
		logger.Info("sentry for operator was initialized")

		// Flush buffered events before the program terminates.
		defer sentry.Flush(2 * time.Second)
	}

	if err := checkEnv(setupLog); err != nil {
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

	// create controller for KuberLogicServices resource
	if err = (&controllers.KuberLogicServiceReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controller").WithName("KuberLogicServices"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KuberLogicServices")
		os.Exit(1)
	}

	// create controller for KuberLogicBackupSchedule resource
	if err = (&controllers.KuberLogicBackupScheduleReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controller-backup").WithName("KuberLogicBackupSchedule"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create backup controller",
			"controller-backup", "KuberLogicBackupSchedule")
		os.Exit(1)
	}

	// create controller for KuberLogicBackupRestore resource
	if err = (&controllers.KuberLogicBackupRestoreReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controller-backup").WithName("KuberLogicBackupSchedule"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create restore controller",
			"controller-restore-backup", "KuberLogicBackupRestore")
		os.Exit(1)
	}

	// init monitoring collector
	klCollector := monitoring.KuberLogicCollector{}
	metrics.Registry.MustRegister(klCollector)

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
