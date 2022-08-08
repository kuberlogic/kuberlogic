/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"flag"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/zapr"
	sentry2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/sentry"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	"os"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	kuberlogicservice_env "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers/kuberlogicservice-env"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(kuberlogiccomv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "manager-config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")
	flag.Parse()

	cfg, err := cfg2.NewConfig()
	if err != nil {
		setupLog.Error(err, "unable to get required config")
		os.Exit(1)
	}

	rawLogger := zap.NewRaw(zap.UseFlagOptions(&zap.Options{
		Development: true,
	}))

	// init sentry
	if dsn := cfg.SentryDsn; dsn != "" {
		err := sentry2.InitSentry(dsn, "operator")
		if err != nil {
			setupLog.Error(err, "unable to init sentry")
			os.Exit(1)
		}
		ctrl.SetLogger(zapr.NewLogger(sentry2.UseSentryWithLogger(dsn, rawLogger, "operator")))

		defer sentry.Flush(2 * time.Second)

		setupLog.Info("sentry for operator is initialized")
	} else {
		setupLog.Info("sentry for operator is not specified")
		ctrl.SetLogger(zapr.NewLogger(rawLogger))
	}

	options := ctrl.Options{
		Scheme:                  scheme,
		LeaderElectionNamespace: cfg.Namespace,
	}
	if configFile != "" {
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile))
		if err != nil {
			setupLog.Error(err, "unable to load the manager config file")
			os.Exit(1)
		}
		setupLog.Info("loaded manager config file", "file", configFile)
	} else {
		setupLog.Info("manager config file is not specified")
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginInstances := make(map[string]commons.PluginService)
	for _, item := range cfg.Plugins {

		// We're a host! Start by launching the plugin process.
		pluginClient := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: commons.HandshakeConfig,
			Plugins: map[string]plugin.Plugin{
				item.Name: &commons.Plugin{},
			},
			Cmd:    exec.Command(item.Path),
			Logger: logger,
		})
		defer func(cl *plugin.Client) {
			cl.Kill()
		}(pluginClient)

		// Connect via RPC
		rpcClient, err := pluginClient.Client()
		if err != nil {
			setupLog.Error(err, "unable connecting to plugin")
			os.Exit(1)
		}

		// Request the plugin
		raw, err := rpcClient.Dispense(item.Name)
		if err != nil {
			setupLog.Error(err, "unable requesting to plugin")
			os.Exit(1)
		}

		pluginInstances[item.Name] = raw.(commons.PluginService)
	}

	// registering watchers for the dependent resources
	var dependantObjects []client.Object
	for pluginType, instance := range pluginInstances {
		setupLog.Info("adding to register watcher", "type", pluginType)
		for _, o := range instance.Types().Objects {
			dependantObjects = append(dependantObjects, o)
		}
	}

	err = (&controllers.KuberLogicServiceReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Plugins: pluginInstances,
		Cfg:     cfg,
	}).SetupWithManager(mgr, dependantObjects...)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KuberLogicService")
		os.Exit(1)
	}

	if err = (&kuberlogiccomv1alpha1.KuberLogicService{}).SetupWebhookWithManager(mgr, pluginInstances); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KuberLogicService")
		os.Exit(1)
	}

	mgr.GetWebhookServer().Register("/mutate-service-pod", &webhook.Admission{
		Handler: &kuberlogicservice_env.ServicePodWebhook{Client: mgr.GetClient()}})

	if cfg.Backups.Enabled {
		setupLog.Info("Backups/Restores are enabled")
		utilruntime.Must(velero.AddToScheme(scheme))

		if err = (&controllers.KuberlogicServiceBackupReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Cfg:    cfg,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KuberlogicServiceBackup")
			os.Exit(1)
		}
		if err = (&controllers.KuberlogicServiceRestoreReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Cfg:    cfg,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KuberlogicServiceRestore")
			os.Exit(1)
		}
		if err = (&controllers.KuberlogicServiceBackupScheduleReconciler{
			Client: mgr.GetClient(),
			Scheme: mgr.GetScheme(),
			Cfg:    cfg,
		}).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KuberlogicServiceBackupSchedule")
			os.Exit(1)
		}
	} else {
		setupLog.Info("Backups/Restores are disabled")
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
