package helm_installer

import (
	"github.com/kuberlogic/operator/modules/installer/internal"
	"github.com/pkg/errors"
)

func (i *HelmInstaller) Upgrade(args []string) error {
	i.Log.Debugf("entering upgrade phase with args: %+v", args)

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

	// for now we only expect single arg = see cmd/install.go
	upgradePhase := args[0]

	err = func() error {
		// upgrade CRDs into cluster
		i.Log.Infof("Upgrading CRDs")
		if err := deployCRDs(globalValues, i); err != nil {
			return errors.Wrap(err, "error upgrading CRDs")
		}

		if upgradePhase == "all" || upgradePhase == "dependencies" {
			i.Log.Infof("Upgrading dependencies...")
			if err := deployNginxIC(globalValues, i, release); err != nil {
				return errors.Wrap(err, "error upgrade nginx-ingress-controller")
			}
			i.Log.Infof("Upgrading cert-manager dependency")
			if err := deployCertManager(globalValues, i); err != nil {
				return errors.Wrap(err, "error installing cert-manager")
			}

			i.Log.Infof("Upgrading authentication component")
			if err := deployAuth(globalValues, i); err != nil {
				return errors.Wrap(err, "error upgrading keycloak")
			}

			i.Log.Infof("Upgrading service operators")
			if err := deployServiceOperators(globalValues, i); err != nil {
				return errors.Wrap(err, "error upgrading service operators")
			}

			i.Log.Infof("Upgrading monitoring component")
			if err := deployMonitoring(globalValues, i); err != nil {
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

	if release.ShowBanner() {
		i.Log.Infof(release.Banner())
	}

	i.Log.Infof("Upgrade completed successfully!")
	return nil
}
