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

package kli

import (
	"github.com/kuberlogic/kuberlogic/modules/installer/cfg"
	helm_installer "github.com/kuberlogic/kuberlogic/modules/installer/kli/helm-installer"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
)

type KuberlogicInstaller interface {
	Install(args []string) error
	Upgrade(args []string) error
	Uninstall(args []string) error
	Status(args []string) error
	Exit(err error)
}

func NewInstaller(config *cfg.Config, log logger.Logger) (KuberlogicInstaller, error) {
	helm, err := helm_installer.New(config, log)
	if err != nil {
		return nil, err
	}
	return helm, nil
}
