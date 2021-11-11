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
)

func (i *HelmInstaller) Uninstall(args []string) error {
	force := false
	for _, a := range args {
		if a == "force" {
			force = true
		}
	}

	i.Log.Debugf("entering uninstall phase with args: %+v", args)
	err := func() error {
		i.Log.Infof("uninstalling core components")
		for _, c := range []string{helmApiserverChart, helmOperatorChart, helmUIChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error uninstalling "+c)
			}
		}

		i.Log.Infof("uninstalling platform components")
		if err := uninstallHelmChart(helmKuberlogicKeycloakCHart, force, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error uninstalling Kuberlogic Keycloak")
		}
		if err := uninstallHelmChart(helmCertManagerChart, force, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error uninstalling cert-manager")
		}

		for _, c := range []string{helmKuberlogicIngressChart, helmKongIngressControllerChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error uninstalling nginx-ingress-controller")
			}
		}

		for _, c := range []string{helmKeycloakOperatorChart, helmMonitoringChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error uninstalling "+c)
			}
		}

		i.Log.Infof("uninstalling service components")
		for _, c := range []string{mysqlOperatorChart, postgresOperatorChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error uninstalling "+c)
			}
		}

		i.Log.Infof("cleaning up environment")
		if err := internal.CleanupEnvironment(i.ReleaseNamespace, i.ClientSet); err != nil {
			return errors.Wrap(err, "error cleaning up the environment")
		}
		if err := internal.CleanupReleaseInfo(i.ReleaseNamespace, i.ClientSet); err != nil {
			return errors.Wrap(err, "error cleaning up release information")
		}
		return nil
	}()
	if err != nil {
		i.Log.Errorf("Uninstall operation failed")
		return err
	}
	i.Log.Infof("Kuberlogic was uninstalled")
	return nil
}
