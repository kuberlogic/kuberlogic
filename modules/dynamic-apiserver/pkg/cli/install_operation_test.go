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
	defaultArgs = []string{
		"--non-interactive",
		"--storage_class", "standard",
		"--ingress_class", "demo",
		"--kuberlogic_domain", "example.com",
	}
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
				Name: "standard",
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

	configFile := filepath.Join(configDir, "config.yaml")
	kubectlBin = "echo"

	if _, err := os.Stat(configFile); err == nil {
		t.Fatalf("config file %s should not exist before install", configFile)
	}

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}
	cmd.SetArgs(append([]string{"install", "--config", configFile}, defaultArgs...))
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(configFile); err != nil {
		t.Fatalf("config file %s should exist after install", configFile)
	}

	if viper.ConfigFileUsed() != configFile {
		t.Fatal("viper should be using created config file")
	}

	if err := viper.ReadInConfig(); err != nil {
		t.Fatal(err)
	}

	if current := viper.GetString("hostname"); current != "127.0.0.1:80" {
		t.Fatal("incorrect hostname: " + current)
	}

	// validate provided options
	klParams := viper.New()
	klParams.SetConfigFile(filepath.Join(configDir, "cache/config/manager/kuberlogic-config.env"))
	if err := klParams.ReadInConfig(); err != nil {
		t.Fatal(err)
	}
	if v := klParams.GetString("INGRESS_CLASS"); v != "demo" {
		t.Fatalf("incorrect ingress class. expected %s, got %s", "demo", v)
	}
}

func TestInstallClusterNotAvailable(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile := filepath.Join(configDir, "config.yaml")
	kubectlBin = "exit 1"

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}

	cmd.SetArgs(append([]string{"install", "--config", configFile}, defaultArgs...))
	if err := cmd.Execute(); err.Error() != "kubernetes is not available via kubectl" {
		t.Fatal(err)
	}
}

func TestIngressClassNotAvailable(t *testing.T) {
	configDir, _ := os.MkdirTemp("", "install-test")
	defer os.RemoveAll(configDir)

	configFile := filepath.Join(configDir, "config.yaml")
	kubectlBin = "echo"

	cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
	if err != nil {
		t.Fatal(err)
	}

	cmd.SetArgs(append([]string{"install", "--config", configFile}, append(defaultArgs, "--ingress_class", "fake")...))
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

			configFile := filepath.Join(configDir, "config.yaml")
			kubectlBin = "echo"

			cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(append([]string{"install", "--config", configFile, "--sentry_dsn", tt.in}, defaultArgs...))
			err = cmd.Execute()
			if err == nil && tt.out == nil {
				return
			}
			if err.Error() != tt.out.Error() {
				t.Errorf("got %q, want %q", err, tt.out)
			}
		})
	}
}

func TestValidationDockerComposeProvided(t *testing.T) {
	testFile, _ := os.CreateTemp(".", "docker-compose")
	testFile.Close()
	defer os.Remove(testFile.Name())

	testDir, _ := os.MkdirTemp(".", "docker-compose")
	defer os.RemoveAll(testDir)

	tests := map[string]error{
		testFile.Name(): nil,
		testDir:         errDirFound,
		"fake123":       os.ErrNotExist,
	}

	for in, expectedErr := range tests {
		t.Run(in, func(t *testing.T) {
			configDir, _ := os.MkdirTemp("", "install-test")
			defer os.RemoveAll(configDir)

			configFile := filepath.Join(configDir, "config.yaml")
			kubectlBin = "echo"

			cmd, err := MakeRootCmd(nil, k8stesting.NewSimpleClientset(fakeClusterResources...))
			if err != nil {
				t.Fatal(err)
			}
			cmd.SetArgs(append([]string{"install", "--config", configFile, "--docker_compose", in}, defaultArgs...))
			err = cmd.Execute()
			if err == nil && expectedErr == nil {
				return
			}
			if !errors.Is(err, expectedErr) {
				t.Fatalf("got %v, expected %v", err, expectedErr)
			}
		})
	}
}
