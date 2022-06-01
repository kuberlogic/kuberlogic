package kuberlogicservice_env

import (
	"context"
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
	client2 "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("KuberlogicserviceEnv Manager", func() {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))

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
			envMgr, err := SetupEnv(kls, client, cfg, context.TODO())
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

		It("Should expose Ingress service", func() {
			envMgr, err := SetupEnv(kls, client, cfg, context.TODO())
			Expect(err).Should(BeNil())

			svc := &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo",
					Namespace: "demo",
				},
				Spec: v1.ServiceSpec{
					Ports: []v1.ServicePort{
						{
							Name:     "demo",
							Protocol: "TCP",
							Port:     80,
						},
					},
				},
			}
			err = client.Create(context.TODO(), svc)
			Expect(err).Should(BeNil())

			By("Setting domain")
			kls.Spec.Domain = "example.com"
			endpoint, err := envMgr.ExposeService(svc.GetName(), true)
			Expect(err).Should(BeNil())
			Expect(endpoint).Should(Equal("http://demo.example.com"))

			By("Enabling HTTPS with default Ingress Certificate")
			kls.Spec.TLSEnabled = true
			endpoint, err = envMgr.ExposeService(svc.GetName(), true)
			Expect(err).Should(BeNil())
			Expect(endpoint).Should(Equal("https://demo.example.com"))
			ingress := &v12.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      kls.GetName(),
					Namespace: kls.Status.Namespace,
				},
			}
			err = client.Get(context.TODO(), client2.ObjectKeyFromObject(ingress), ingress)
			Expect(err).Should(BeNil())
			Expect(ingress.Spec.TLS[0].SecretName).Should(Equal(""))
			Expect(ingress.Spec.TLS[0].Hosts[0]).Should(Equal("demo.example.com"))

			By("Enabling HTTPS with custom Ingress Certificate")
			tlsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "demo",
					Namespace: cfg.Namespace,
				},
				Data: map[string][]byte{
					"tls.key": []byte("test"),
					"tls.crt": []byte("demo"),
				},
			}
			Expect(client.Create(context.TODO(), tlsSecret)).Should(BeNil())

			cfg.SvcOpts.TLSSecretName = tlsSecret.GetName()
			envMgr, err = SetupEnv(kls, client, cfg, context.TODO())
			Expect(err).Should(BeNil())

			nsSecret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cfg.SvcOpts.TLSSecretName,
					Namespace: envMgr.NamespaceName,
				},
			}
			Expect(client.Get(context.TODO(), client2.ObjectKeyFromObject(nsSecret), nsSecret)).Should(BeNil())
			Expect(nsSecret.Data).Should(Equal(tlsSecret.Data))

			endpoint, err = envMgr.ExposeService(svc.GetName(), true)
			Expect(err).Should(BeNil())
			Expect(endpoint).Should(Equal("https://demo.example.com"))
			err = client.Get(context.TODO(), client2.ObjectKeyFromObject(ingress), ingress)
			Expect(err).Should(BeNil())
			Expect(ingress.Spec.TLS[0].SecretName).Should(Equal(nsSecret.GetName()))
			Expect(ingress.Spec.TLS[0].Hosts[0]).Should(Equal("demo.example.com"))
		})
	})
})
