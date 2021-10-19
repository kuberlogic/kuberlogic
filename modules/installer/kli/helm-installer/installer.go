/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package helm_installer

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/installer/cfg"
	kubeConfig "github.com/kuberlogic/kuberlogic/modules/installer/kubernetes"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/client-go/kubernetes"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"os"
)

const (
	helmMaxHistory = 10
)

type HelmInstaller struct {
	Log              logger.Logger
	ClientSet        *kubernetes.Clientset
	HelmActionConfig *action.Configuration
	Config           cfg.Config
	ReleaseNamespace string
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
		Log:              log,
		ClientSet:        k8sclientset,
		HelmActionConfig: helmActionConfig,
		Config:           *config,
		ReleaseNamespace: *config.Namespace,

	}

	return i, nil
}
