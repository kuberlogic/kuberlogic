package helm_installer

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/installer/cfg"
	kubeConfig "github.com/kuberlogic/kuberlogic/modules/installer/kubernetes"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"os"
)

const (
	helmMaxHistory = 10
)

type HelmInstaller struct {
	Log logger.Logger

	ClientSet        *kubernetes.Clientset
	HelmActionConfig *action.Configuration

	ReleaseNamespace string
	Registry         struct {
		Server   string
		Username string
		Password string
	}
	Endpoints struct {
		API               string
		UI                string
		MonitoringConsole string
	}
	Auth struct {
		AdminPassword    string
		TestUserPassword string
	}
}

func New(config *cfg.Config, log logger.Logger) (*HelmInstaller, error) {
	log.Debugf("initializing clientset with kubeconfig %s", *config.KubeconfigPath)
	k8sclientset, err := kubeConfig.NewKubeClient(*config.KubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("error building Kubernetes clientset: %v", err)
	}

	// helm still expects namespace to be set via env var or pflag
	if err := os.Setenv("HELM_NAMESPACE", *config.Namespace); err != nil {
		return nil, err
	}

	// helm client with "secret" driver
	settings := cli.New()
	if *config.DebugLogs {
		settings.Debug = true
	}
	settings.KubeConfig = *config.KubeconfigPath
	settings.MaxHistory = helmMaxHistory

	helmActionConfig := new(action.Configuration)
	if err := helmActionConfig.Init(settings.RESTClientGetter(), *config.Namespace, "secret", log.Debugf); err != nil {
		return nil, fmt.Errorf("error building Helm cli: %v", err)
	}

	i := &HelmInstaller{
		Log: log,

		ClientSet:        k8sclientset,
		HelmActionConfig: helmActionConfig,

		ReleaseNamespace: *config.Namespace,
		Registry: struct {
			Server   string
			Username string
			Password string
		}{
			Server:   config.Registry.Server,
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		},
		Endpoints: struct {
			API               string
			UI                string
			MonitoringConsole string
		}{
			API:               config.Endpoints.API,
			UI:                config.Endpoints.UI,
			MonitoringConsole: config.Endpoints.MonitoringConsole,
		},
		Auth: struct {
			AdminPassword    string
			TestUserPassword string
		}{
			AdminPassword: config.Auth.AdminPassword,
		},
	}
	if config.Auth.TestUserPassword != nil {
		i.Auth.TestUserPassword = *config.Auth.TestUserPassword
	}

	return i, nil
}
