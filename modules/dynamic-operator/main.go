/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"os"
	"os/exec"
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

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"postgresql": &commons.Plugin{},
}

func main() {
	//var metricsAddr string
	//var enableLeaderElection bool
	//var probeAddr string
	//flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	//flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	//flag.BoolVar(&enableLeaderElection, "leader-elect", false,
	//	"Enable leader election for controller manager. "+
	//		"Enabling this will ensure there is only one active controller manager.")

	//opts.BindFlags(flag.CommandLine)
	//flag.Parse()
	opts := zap.Options{
		Development: true,
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	cfg, err := cfg2.NewConfig()
	if err != nil {
		setupLog.Error(err, "unable to get required config")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     cfg.MetricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: cfg.ProbeAddr,
		LeaderElection:         cfg.EnableLeaderElection,
		LeaderElectionID:       "afb3d480.kuberlogic.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// init postgresql plugin configuration
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
			Plugins:         pluginMap,
			Cmd:             exec.Command(item.Path),
			Logger:          logger,
		})
		defer pluginClient.Kill()

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
		dependantObjects = append(dependantObjects, instance.Type().Object)
	}

	err = (&controllers.KuberLogicServiceReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Plugins: pluginInstances,
	}).SetupWithManager(mgr, dependantObjects...)
	if err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KuberLogicService")
		os.Exit(1)
	}

	if err = (&kuberlogiccomv1alpha1.KuberLogicService{}).SetupWebhookWithManager(mgr, pluginInstances); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "KuberLogicService")
		os.Exit(1)
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
