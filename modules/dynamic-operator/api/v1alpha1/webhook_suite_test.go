/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"

	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	//+kubebuilder:scaffold:imports
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	testK8sClient client.Client
	testEnv       *envtest.Environment
	ctx           context.Context
	cancel        context.CancelFunc
	pluginClients []*plugin.Client
)

var pluginMap = map[string]plugin.Plugin{
	"docker-compose": &commons.Plugin{},
}

func TestWebhookAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Webhook Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	config, err := cfg2.NewConfig()
	Expect(err).NotTo(HaveOccurred())

	ctx, cancel = context.WithCancel(context.TODO())

	scheme := runtime.NewScheme()
	err = AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1beta1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = corev1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	useExistingCluster := os.Getenv("USE_EXISTING_CLUSTER") == "true"
	if useExistingCluster {
		testEnv = &envtest.Environment{
			UseExistingCluster: &useExistingCluster,
		}
		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		testK8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(testK8sClient).NotTo(BeNil())
	} else {

		By("bootstrapping test environment")
		testEnv = &envtest.Environment{
			CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
			ErrorIfCRDPathMissing: false,
			WebhookInstallOptions: envtest.WebhookInstallOptions{
				Paths: []string{filepath.Join("..", "..", "config", "webhook")},
			},
		}

		cfg, err := testEnv.Start()
		Expect(err).NotTo(HaveOccurred())
		Expect(cfg).NotTo(BeNil())

		testK8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
		Expect(err).NotTo(HaveOccurred())
		Expect(testK8sClient).NotTo(BeNil())

		ns := &corev1.Namespace{}
		ns.SetName(os.Getenv("NAMESPACE"))
		Expect(testK8sClient.Create(ctx, ns)).Should(Succeed())

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

		// start webhook server using Manager
		webhookInstallOptions := &testEnv.WebhookInstallOptions
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:             scheme,
			Host:               webhookInstallOptions.LocalServingHost,
			Port:               webhookInstallOptions.LocalServingPort,
			CertDir:            webhookInstallOptions.LocalServingCertDir,
			LeaderElection:     false,
			MetricsBindAddress: "0",
		})
		Expect(err).NotTo(HaveOccurred())

		err = (&KuberLogicService{}).SetupWebhookWithManager(mgr, pluginInstances)
		Expect(err).NotTo(HaveOccurred())

		err = (&KuberlogicServiceBackup{}).SetupWebhookWithManager(mgr, config.Backups.Enabled)
		Expect(err).NotTo(HaveOccurred())

		//+kubebuilder:scaffold:webhook

		go func() {
			defer GinkgoRecover()
			err = mgr.Start(ctx)
			Expect(err).NotTo(HaveOccurred())
		}()

		// wait for the webhook server to get ready
		dialer := &net.Dialer{Timeout: time.Second}
		addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
		Eventually(func() error {
			conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
			if err != nil {
				return err
			}
			conn.Close()
			return nil
		}).Should(Succeed())
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
