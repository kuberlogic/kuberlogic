/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	velero "github.com/vmware-tanzu/velero/pkg/apis/velero/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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
	"docker-compose": &commons.Plugin{},
}

type fakeExecutor struct {
	err error
}

func (f *fakeExecutor) Stream(o remotecommand.StreamOptions) error {
	return f.err
}

func TestControllerAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")

	err := kuberlogiccomv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = velero.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	if useExistingCluster() {

		err = os.Unsetenv("KUBEBUILDER_ASSETS")
		Expect(err).NotTo(HaveOccurred())

		testEnv = &envtest.Environment{
			UseExistingCluster: pointer.Bool(useExistingCluster()),
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())
	} else {
		testEnv = &envtest.Environment{
			CRDDirectoryPaths: []string{
				filepath.Join("..", "config", "crd", "bases"),
				filepath.Join("..", "config", "crd", "velero"),
			},
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(k8sClient).NotTo(BeNil())

		config, err := cfg2.NewConfig()
		Expect(err).NotTo(HaveOccurred())

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
			Client:     k8sManager.GetClient(),
			Scheme:     k8sManager.GetScheme(),
			RESTConfig: cfg,
			Plugins:    pluginInstances,
			Cfg:        config,
		}).SetupWithManager(k8sManager, dependantObjects...)
		Expect(err).ToNot(HaveOccurred())

		if config.Backups.Enabled {
			ns := &corev1.Namespace{}
			ns.SetName("velero")
			Expect(k8sClient.Create(ctx, ns)).Should(Succeed())

			err = (&KuberlogicServiceBackupReconciler{
				Client: k8sManager.GetClient(),
				Scheme: k8sManager.GetScheme(),
				Cfg:    config,
			}).SetupWithManager(k8sManager)
			Expect(err).ToNot(HaveOccurred())

			err = (&KuberlogicServiceRestoreReconciler{
				Client: k8sManager.GetClient(),
				Scheme: k8sManager.GetScheme(),
				Cfg:    config,
			}).SetupWithManager(k8sManager)
			Expect(err).ToNot(HaveOccurred())

			err = (&KuberlogicServiceBackupScheduleReconciler{
				Client: k8sManager.GetClient(),
				Scheme: k8sManager.GetScheme(),
				Cfg:    config,
			}).SetupWithManager(k8sManager)
			Expect(err).ToNot(HaveOccurred())
		}

		go func() {
			defer GinkgoRecover()
			err = k8sManager.Start(ctx)
			Expect(err).ToNot(HaveOccurred(), "failed to run manager")
		}()
	}
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

func useExistingCluster() bool {
	return os.Getenv("USE_EXISTING_CLUSTER") == "true"
}
