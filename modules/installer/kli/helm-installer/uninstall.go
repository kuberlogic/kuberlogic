package helm_installer

import (
	"github.com/kuberlogic/operator/modules/installer/internal"
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
		for _, c := range []string{helmApiserverChart, helmOperatorChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return err
			}
		}
		i.Log.Infof("uninstalling service components")
		for _, c := range []string{mysqlOperatorChart, postgresOperatorChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return err
			}
		}

		i.Log.Infof("uninstalling platform components")
		if err := uninstallKuberlogicKeycloak(i.ReleaseNamespace, force, i.HelmActionConfig, i.ClientSet, i.Log); err != nil {
			return err
		}

		for _, c := range []string{helmKeycloakOperatorChart, helmMonitoringChart} {
			if err := uninstallHelmChart(c, force, i.HelmActionConfig, i.Log); err != nil {
				return err
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
