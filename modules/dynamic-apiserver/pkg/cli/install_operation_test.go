package cli

import (
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	k8stesting "k8s.io/client-go/kubernetes/fake"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallLB(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile = filepath.Join(configDir, "config.yaml")
	kubectlBin = "echo"
	kustomizeBin = "echo"
	viper.SetConfigFile(configFile)

	cmd := makeInstallCmd(func() (kubernetes.Interface, error) {
		return k8stesting.NewSimpleClientset(&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kls-api-server",
				Namespace: "kuberlogic",
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Port: 80,
					},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP: "127.0.0.1",
						},
					},
				},
			},
		}), nil
	})
	cmd.SetArgs([]string{"--non-interactive"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	if current := viper.GetString("hostname"); current != "127.0.0.1:80" {
		t.Fatal("incorrect hostname: " + current)
	}
}

func TestInstallClusterNotAvailable(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile = filepath.Join(configDir, "config.yaml")
	kubectlBin = "exit 1"
	kustomizeBin = "exit 1"
	viper.SetConfigFile(configFile)

	cmd := makeInstallCmd(func() (kubernetes.Interface, error) {
		return k8stesting.NewSimpleClientset(), nil
	})
	cmd.SetArgs([]string{"--non-interactive"})
	if err := cmd.Execute(); err.Error() != "Kubernetes is not available via kubectl" {
		t.Fatal(err)
	}
}
