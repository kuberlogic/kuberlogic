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
