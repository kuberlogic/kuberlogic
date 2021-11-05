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

package kuberlogictenant_controller

import (
	"github.com/go-logr/logr"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"github.com/kuberlogic/kuberlogic/modules/operator/controllers/kuberlogictenant_controller/grafana"
)

// finalize function "resolves" an alert when kuberlogictenant is deleted.
func finalize(cfg *cfg.Config, kt *kuberlogicv1.KuberLogicTenant, log logr.Logger) error {
	log.Info("processing finalizer")
	if cfg.Grafana.Enabled {
		log.Info("processing grafana organizations and users")
		if err := grafana.NewGrafanaSyncer(kt, log, cfg.Grafana).CleanupOrg(kt.Name); err != nil {
			return err
		}
	}

	log.Info("checking kuberlogic services")
	if len(kt.Status.Services) != 0 {
		return errFinalizingTenant
	}

	return nil
}
