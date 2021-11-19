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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func (i *HelmInstaller) Upgrade(_ *cobra.Command, args []string) error {
	i.Log.Debugf("entering upgrade phase with args: %+v", args)
	upgradePhase := "all"
	if len(args) > 0 {
		upgradePhase = args[0]
	}
	i.Log.Debugf("using install phase: %s", upgradePhase)

	// check if release is installed
	release, found, err := internal.DiscoverReleaseInfo(i.ReleaseNamespace, i.ClientSet)
	if err != nil {
		return errors.Wrap(err, "error searching for the release")
	}
	if !found {
		return errors.New("release not found")
	}
	if err := release.UpgradeRelease(i.ClientSet); err != nil {
		return errors.Wrap(err, "error starting upgrade")
	}

	err = func() error {
		// do not pass imagePullSecretReference if it is disabled
		if i.Config.Registry.Server == "" {
			delete(globalValues, "imagePullSecrets")
		}

		// upgrade CRDs into cluster
		i.Log.Infof("Upgrading CRDs")
		if err := deployCRDs(globalValues, i); err != nil {
			return errors.Wrap(err, "error upgrading CRDs")
		}

		if upgradePhase == "all" || upgradePhase == "dependencies" {
			i.Log.Infof("Upgrading dependencies...")
			i.Log.Infof("Upgrading cert-manager dependency")
			if err := deployCertManager(globalValues, i); err != nil {
				return errors.Wrap(err, "error installing cert-manager")
			}

			i.Log.Infof("Upgrading authentication component")
			if err := deployAuth(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrading keycloak")
			}
			if err := deployIngressController(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrade nginx-ingress-controller")
			}
			i.Log.Infof("Upgrading service operators")
			if err := deployServiceOperators(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrading service operators")
			}

			i.Log.Infof("Upgrading monitoring component")
			if err := deployMonitoring(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrading monitoring component")
			}
		}

		if upgradePhase == "all" || upgradePhase == "kuberlogic" {
			i.Log.Infof("Upgrading Kuberlogic core components...")
			i.Log.Infof("Upgrading operator")
			if err := deployOperator(globalValues, i); err != nil {
				return errors.Wrap(err, "error upgrading operator")
			}

			i.Log.Infof("Upgrading apiserver")
			if err := deployApiserver(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrading apiserver")
			}

			i.Log.Infof("Upgrading UI")
			if err := deployUI(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrading UI")
			}
		}

		return nil
	}()
	if err != nil {
		i.Log.Errorf("Upgrade failed")
		return err
	}

	i.Log.Infof(release.ShowBanner())

	i.Log.Infof("Upgrade completed successfully!")
	return release.FinishRelease(i.ClientSet)
}
