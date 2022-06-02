/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"os"
	"os/exec"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	k8sClient     client.Client // You'll be using this client in your tests.
	testEnv       *envtest.Environment
	ctx           context.Context
	cancel        context.CancelFunc
	pluginClients []*plugin.Client
)

var pluginMap = map[string]plugin.Plugin{
	"postgresql": &commons.Plugin{},
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	config, err := cfg2.NewConfig()
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = kuberlogiccomv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	// Create an hclog.Logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:   "plugin",
		Output: os.Stdout,
		Level:  hclog.Debug,
	})

	pluginInstances := make(map[string]commons.PluginService)
	for _, item := range config.Plugins {
		// We're a host! Start by launching the plugin process.
		pluginClient := plugin.NewClient(&plugin.ClientConfig{
			HandshakeConfig: commons.HandshakeConfig,
			Plugins:         pluginMap,
			Cmd:             exec.Command(item.Path),
			Logger:          logger,
		})
		pluginClients = append(pluginClients, pluginClient)

		// Connect via RPC
		rpcClient, err := pluginClient.Client()
		Expect(err).ToNot(HaveOccurred())

		// Request the plugin
		raw, err := rpcClient.Dispense(item.Name)
		Expect(err).ToNot(HaveOccurred())

		pluginInstances[item.Name] = raw.(commons.PluginService)
	}

	// registering watchers for the dependent resources
	var dependantObjects []client.Object
	for _, instance := range pluginInstances {
		for _, o := range instance.Types().Objects {
			dependantObjects = append(dependantObjects, o)
		}
	}

	err = (&KuberLogicServiceReconciler{
		Client:  k8sManager.GetClient(),
		Scheme:  k8sManager.GetScheme(),
		Plugins: pluginInstances,
	}).SetupWithManager(k8sManager, dependantObjects...)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	for _, cl := range pluginClients {
		cl.Kill()
	}

	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
