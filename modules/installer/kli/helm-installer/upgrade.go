package helm_installer

import (
	"fmt"
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
	installPhase := args[0]

	// set release state to upgrading
	if _, err := internal.UpgradeRelease(i.ReleaseNamespace, i.ClientSet); err != nil {
		return errors.Wrap(err, "error starting release")
	}

	// install CRDs into cluster
	i.Log.Infof("Upgrading CRDs")
	if err := deployCRDs(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error upgrading CRDs")
	}

	if installPhase == "all" || installPhase == "dependencies" {
		i.Log.Infof("Upgrading authentication component")
		if err := deployAuth(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
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

	if installPhase == "all" || installPhase == "kuberlogic" {
		i.Log.Infof("Upgrading operator")
		if err := deployOperator(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error upgrading operator")
		}

		i.Log.Infof("Upgrading apiserver")
		if err := deployApiserver(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error upgrading apiserver")
		}
	}

	i.Log.Infof("Installation completed successfully!")

	return fmt.Errorf("not implemented")
}