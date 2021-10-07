package helm_installer

import (
	"github.com/kuberlogic/kuberlogic/modules/installer/internal"
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
	i.Log.Infof(release.ShowBanner())

	return nil
}
