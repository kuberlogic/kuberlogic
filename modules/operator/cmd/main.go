package cmd

import (
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	operatorConfig "github.com/kuberlogic/operator/modules/operator/cfg"
	"github.com/kuberlogic/operator/modules/operator/controllers"
	"github.com/kuberlogic/operator/modules/operator/logging"
	"github.com/kuberlogic/operator/modules/operator/monitoring"
	"github.com/kuberlogic/operator/modules/operator/notifications"
	"github.com/kuberlogic/operator/modules/operator/util"
	mysql "github.com/presslabs/mysql-operator/pkg/apis"
	postgres "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
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
	// +kubebuilder:scaffold:scheme
}

func Main(args []string) {
	cfg, err := operatorConfig.NewConfig()
	if err != nil {
		setupLog.Error(err, "unable to get required config")
		os.Exit(1)
	}
	// populate some values that are used later on
	util.InitFromConfig(cfg)

	zapl, err := logging.CreateZapLogger()
	if err != nil {
		setupLog.Error(err, "unable to create logger")
		os.Exit(1)
	}
	logger := logging.GetLogger(zapl)

	// init sentry
	if dsn := cfg.SentryDsn; dsn != "" {
		logger = logging.UseSentry(dsn, zapl)
		setupLog.Info("sentry for operator was initialized")
	}
	ctrl.SetLogger(logger)

	metricsAddr := cfg.MetricsAddr
	enableLeaderElection := cfg.EnableLeaderElection
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

	// init monitoring collector
	klCollector := monitoring.NewCollector()
	metrics.Registry.MustRegister(klCollector)

	// create controller for KuberLogicServices resource
	if err = (&controllers.KuberLogicServiceReconciler{
		Client:              mgr.GetClient(),
		Log:                 ctrl.Log.WithName("controller").WithName("KuberLogicServices"),
		Scheme:              mgr.GetScheme(),
		MonitoringCollector: klCollector,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KuberLogicServices")
		os.Exit(1)
	}

	// create controller for KuberLogicBackupSchedule resource
	if err = (&controllers.KuberLogicBackupScheduleReconciler{
		Client:              mgr.GetClient(),
		Log:                 ctrl.Log.WithName("controller-backup").WithName("KuberLogicBackupSchedule"),
		Scheme:              mgr.GetScheme(),
		MonitoringCollector: klCollector,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create backup controller",
			"controller-backup", "KuberLogicBackupSchedule")
		os.Exit(1)
	}

	// create controller for KuberLogicBackupRestore resource
	if err = (&controllers.KuberLogicBackupRestoreReconciler{
		Client:              mgr.GetClient(),
		Log:                 ctrl.Log.WithName("controller-backup").WithName("KuberLogicBackupSchedule"),
		Scheme:              mgr.GetScheme(),
		MonitoringCollector: klCollector,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create restore controller",
			"controller-restore-backup", "KuberLogicBackupRestore")
		os.Exit(1)
	}

	// init notification manager
	notifMgr := notifications.NewWithConfig(cfg)
	// create controller for KuberlogicAlert resource
	if err = (&controllers.KuberLogicAlertReconciler{
		Client:               mgr.GetClient(),
		Log:                  ctrl.Log.WithName("controller-alert").WithName("KuberlogicAlert"),
		Scheme:               mgr.GetScheme(),
		NotificationsManager: notifMgr,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create alert controller",
			"controller-alert", "KuberlogicAlert")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
