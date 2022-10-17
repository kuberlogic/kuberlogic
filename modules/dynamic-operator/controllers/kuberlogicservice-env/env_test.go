package kuberlogicservice_env

import (
	"context"

	certmanagerv1 "github.com/jetstack/cert-manager/pkg/apis/certmanager/v1"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	cfg2 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("KuberlogicserviceEnv Manager", func() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(certmanagerv1.AddToScheme(scheme))

	b := fake.NewClientBuilder()

	Context("When creating Kuberlogicservice", func() {
		client := b.WithScheme(scheme).Build()
		cfg := &cfg2.Config{
			Namespace: "kuberlogic",
			SvcOpts: struct {
				TLSSecretName string `envconfig:"optional"`
			}{
				TLSSecretName: "",
			},
		}

		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
			Spec:   v1alpha1.KuberLogicServiceSpec{},
			Status: v1alpha1.KuberLogicServiceStatus{},
		}

		It("Should prepare environment", func() {
			envMgr := New(client, kls, cfg)
			err := envMgr.SetupEnv(context.TODO())
			Expect(err).Should(BeNil())
			Expect(envMgr.NamespaceName).Should(Equal(kls.Name))

			By("Checking created objects")
			ns := &v1.Namespace{}
			err = client.Get(context.TODO(), types.NamespacedName{Name: envMgr.NamespaceName}, ns)
			Expect(err).Should(BeNil())
			Expect(ns.GetName()).Should(Equal(envMgr.NamespaceName))

			netpol := &v12.NetworkPolicyList{}
			err = client.List(context.TODO(), netpol)
			Expect(err).Should(BeNil())
			Expect(len(netpol.Items)).Should(Equal(1))
		})
	})

	Context("When creating Kuberlogicservice with docker-registry credentials", func() {
		client := b.WithScheme(scheme).Build()
		cfg := &cfg2.Config{
			Namespace: "kuberlogic",
			DockerRegistry: struct {
				Url      string
				Username string
				Password string
			}{
				Url:      "docker.io",
				Username: "test",
				Password: "test",
			},
			SvcOpts: struct {
				TLSSecretName string `envconfig:"optional"`
			}{
				TLSSecretName: "",
			},
		}

		kls := &v1alpha1.KuberLogicService{
			ObjectMeta: metav1.ObjectMeta{
				Name: "docker-registry-demo",
			},
			Spec:   v1alpha1.KuberLogicServiceSpec{},
			Status: v1alpha1.KuberLogicServiceStatus{},
		}

		It("Should prepare environment", func() {
			envMgr := New(client, kls, cfg)
			err := envMgr.SetupEnv(context.TODO())
			Expect(err).Should(BeNil())
			Expect(envMgr.NamespaceName).Should(Equal(kls.Name))

			By("Checking created objects")
			ns := &v1.Namespace{}
			err = client.Get(context.TODO(), types.NamespacedName{Name: envMgr.NamespaceName}, ns)
			Expect(err).Should(BeNil())
			Expect(ns.GetName()).Should(Equal(envMgr.NamespaceName))

			netpol := &v12.NetworkPolicyList{}
			err = client.List(context.TODO(), netpol)
			Expect(err).Should(BeNil())
			Expect(len(netpol.Items)).Should(Equal(1))

			secretList := &v1.SecretList{}
			err = client.List(context.TODO(), secretList)
			Expect(err).Should(BeNil())
			Expect(len(secretList.Items)).Should(Equal(1))
			Expect(secretList.Items[0].Name).Should(Equal("docker-registry"))
			dockerJson, err := dockerCredsToJson(
				cfg.DockerRegistry.Url,
				cfg.DockerRegistry.Username,
				cfg.DockerRegistry.Password,
			)
			Expect(err).Should(BeNil())
			Expect(secretList.Items[0].Data).Should(Equal(map[string][]byte{
				".dockerconfigjson": dockerJson,
			}))
		})
	})
})
