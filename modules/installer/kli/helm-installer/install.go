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
	"github.com/kuberlogic/kuberlogic/modules/installer/internal"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
)

func (i *HelmInstaller) Install(args []string) error {
	i.Log.Debugf("entering install phase with args: %+v", args)

	// for now we only expect single arg = see cmd/install.go
	installPhase := args[0]

	// run pre install checks
	if err := runInstallChecks(i.ClientSet, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "pre-install checks are failed")
	}

	// prepare environment for release and start release process
	if err := internal.PrepareEnvironment(i.ReleaseNamespace, i.Registry.Server, i.Registry.Password, i.Registry.Username, i.ClientSet); err != nil {
		return errors.Wrap(err, "error preparing environment")
	}
	release, err := internal.StartRelease(i.ReleaseNamespace, i.ClientSet)
	if err != nil {
		return errors.Wrap(err, "error starting release")
	}

	err = func() error {
		// do not pass imagePullSecretReference if it is disabled
		if i.Registry.Server == "" {
			delete(globalValues, "imagePullSecrets")
		}

		// install CRDs into cluster
		i.Log.Infof("Installing CRDs...")
		if err := deployCRDs(globalValues, i); err != nil {
			return errors.Wrap(err, "error installing CRDs")
		}

		if installPhase == "all" || installPhase == "dependencies" {
			i.Log.Infof("Installing Kuberlogic dependencies...")
			if err := deployCertManager(globalValues, i); err != nil {
				return errors.Wrap(err, "error installing cert-manager")
			}

			if err := deployAuth(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing keycloak")
			}
			if err := deployIngressController(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing nginx-ingress-controller")
			}
			if err := deployServiceOperators(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing service operators")
			}

			if err := deployMonitoring(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing monitoring component")
			}
		}

		if installPhase == "all" || installPhase == "kuberlogic" {
			i.Log.Infof("Installing Kuberlogic core components...")
			if err := deployOperator(globalValues, i); err != nil {
				return errors.Wrap(err, "error installing operator")
			}

			if err := deployApiserver(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing apiserver")
			}

			if err := deployUI(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error installing UI")
			}
		}
		return nil
	}()
	if err != nil {
		i.Log.Infof("Installation failed: %v", err)
		internal.FailRelease(i.ReleaseNamespace, i.ClientSet)
		return err
	}
	i.Log.Infof(release.ShowBanner())

	err = release.FinishRelease(i.ClientSet)
	i.Log.Infof("Installation completed successfully!")
	return err
}

func runInstallChecks(clientSet *kubernetes.Clientset, actionConfig *action.Configuration, log logger.Logger) error {
	if err := checkKubernetesVersion(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes version")
	}
	if err := checkDefaultStorageClass(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes default StorageClass")
	}
	if err := checkLoadBalancerServiceType(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes LoadBalancer service")
	}
	return nil
}
