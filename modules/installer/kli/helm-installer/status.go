package helm_installer

import (
	"github.com/kuberlogic/operator/modules/installer/internal"
)

func (i *HelmInstaller) Status(args []string) error {
	i.Log.Debugf("entering status phase with args: %+v", args)

	release, installed, err := internal.DiscoverReleaseInfo(i.ReleaseNamespace, i.ClientSet)
	if err != nil {
		i.Log.Fatalf("Error searching for release: %s", err.Error())
		return nil
	}
	if !installed {
		i.Log.Infof("Kuberlogic release is not found")
		return nil
	}
	i.Log.Infof("Kuberlogic status is: %s", release.Status)

	i.Log.Debugf("release banner: %s", release.Banner())
	if release.ShowBanner() {
		i.Log.Infof(release.Banner())
	}

	return nil
}
