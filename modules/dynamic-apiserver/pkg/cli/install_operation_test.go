package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	v12 "k8s.io/api/networking/v1"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	k8stesting "k8s.io/client-go/kubernetes/fake"
)

var (
	fakeClusterResources = []runtime.Object{
		&corev1.Service{
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
		},
		&v1.StorageClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
		},
		&v12.IngressClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: "demo",
			},
		},
	}
)

func TestInstallLB(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile = filepath.Join(configDir, "config.yaml")
	kubectlBin = "echo"
	viper.SetConfigFile(configFile)

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"install", "--non-interactive", "--storage_class", "demo", "--ingress_class", "demo"})
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
	viper.SetConfigFile(configFile)

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"install", "--non-interactive", "--storage_class", "demo", "--ingress_class", "demo"})
	if err := cmd.Execute(); err.Error() != "Kubernetes is not available via kubectl" {
		t.Fatal(err)
	}
}

func TestIngressClassNotAvailable(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile = filepath.Join(configDir, "config.yaml")
	kubectlBin = "exit 1"
	viper.SetConfigFile(configFile)

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs([]string{"install", "--non-interactive", "--storage_class", "demo", "--ingress_class", "fake"})
	if err := cmd.Execute(); err.Error() != "error processing ingress_class flag: fake is not available. Available: demo" {
		t.Fatal(err)
	}
}

func TestValidationSentryURI(t *testing.T) {
	tests := []struct {
		in  string
		out error
	}{
		{"", nil},
		{"https://b16abaff497941468fdf21aff686ff52@kl.sentry.cloudlinux.com/9", nil},
		{"N", errors.Wrapf(errors.Wrap(errSentryInvalidURI, "N"), "error processing %s flag", installSentryDSNParam)},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			configDir, _ := os.MkdirTemp("", "install-test")
			defer os.RemoveAll(configDir)

			configFile = filepath.Join(configDir, "config.yaml")
			kubectlBin = "echo"
			viper.SetConfigFile(configFile)

			cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs([]string{"install", "--non-interactive", "--storage_class", "demo", "--ingress_class", "demo", "--sentry_dsn", tt.in})
			err = cmd.Execute()
			if err == nil && tt.out == nil {
				return
			}
			if tt.out.Error() != err.Error() {
				t.Errorf("got %q, want %q", err, tt.out)
			}
		})
	}
}
