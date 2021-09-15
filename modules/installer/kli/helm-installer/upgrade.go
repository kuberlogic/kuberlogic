package helm_installer

import (
	"github.com/kuberlogic/operator/modules/installer/internal"
	"github.com/pkg/errors"
)

func (i *HelmInstaller) Upgrade(args []string) error {
	i.Log.Debugf("entering upgrade phase with args: %+v", args)

	// check if release is installed
	_, err := internal.UpgradeRelease(i.ReleaseNamespace, i.ClientSet)
	if err != nil {
		i.Log.Errorf("Error searching for the release: %v", err)
		return err
	}

	// for now we only expect single arg = see cmd/install.go
	upgradePhase := args[0]

	// set release state to upgrading
	if _, err := internal.UpgradeRelease(i.ReleaseNamespace, i.ClientSet); err != nil {
		return errors.Wrap(err, "error starting upgrade")
	}

	err = func() error {
		// upgrade CRDs into cluster
		i.Log.Infof("Upgrading CRDs")
		if err := deployCRDs(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error upgrading CRDs")
		}

		if upgradePhase == "all" || upgradePhase == "dependencies" {
			i.Log.Infof("Upgrading dependencies...")
			if err := deployNginxIC(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.ClientSet, i.Log); err != nil {
				return errors.Wrap(err, "error upgrade nginx-ingress-controller")
			}
			i.Log.Infof("Upgrading cert-manager dependency")
			if err := deployCertManager(globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing cert-manager")
			}

			i.Log.Infof("Upgrading authentication component")
			if err := deployAuth(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log, i.ClientSet); err != nil {
				return errors.Wrap(err, "error upgrading keycloak")
			}

			i.Log.Infof("Upgrading service operators")
			if err := deployServiceOperators(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error upgrading service operators")
			}

			i.Log.Infof("Upgrading monitoring component")
			if err := deployMonitoring(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error upgrading monitoring component")
			}
		}

		if upgradePhase == "all" || upgradePhase == "kuberlogic" {
			i.Log.Infof("Upgrading Kuberlogic core components...")
			i.Log.Infof("Upgrading operator")
			if err := deployOperator(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error upgrading operator")
			}

			i.Log.Infof("Upgrading apiserver")
			if err := deployApiserver(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error upgrading apiserver")
			}
		}

		return nil
	}()
	if err != nil {
		i.Log.Errorf("Upgrade failed")
		return err
	}

	i.Log.Infof("Upgrade completed successfully!")
	return nil
}
