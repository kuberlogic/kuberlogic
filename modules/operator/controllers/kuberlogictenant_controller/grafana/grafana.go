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

package grafana

import (
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"github.com/pkg/errors"
)

type grafana struct {
	kt *v1.KuberLogicTenant

	datasourceAddress string
	api               *API
	log               logr.Logger
}

func NewGrafanaSyncer(kt *v1.KuberLogicTenant, log logr.Logger, cfg cfg.Grafana) *grafana {
	return &grafana{
		kt:                kt,
		api:               newGrafanaApi(log, cfg),
		datasourceAddress: cfg.DefaultDatasourceEndpoint,
		log:               log,
	}
}

func (gr *grafana) Sync() error {
	// ensure that there is a dedicated Grafana Organization for this tenant
	orgId, err := gr.ensureOrganization(gr.kt.Name)
	if err != nil {
		return errors.Wrap(err, "error creating grafana organization")
	}
	// ensure that there is a user with Viewer role for this organization
	// this user is used by the Kuberlogic tenant user to access dashboards
	if err := gr.ensureUser(gr.kt.Spec.OwnerEmail, "", uuid.New().String(), VIEWER_ROLE, orgId); err != nil {
		return errors.Wrap(err, "error creating grafana viewer user")
	}
	// ensure that Kuberlogic datasource exists
	if err := gr.ensureDatasource(orgId); err != nil {
		return errors.Wrap(err, "error managing Grafana datasource")
	}
	// create Grafana dashboards
	if err := gr.ensureDashboards(gr.kt.Status.Services, orgId); err != nil {
		return errors.Wrap(err, "error creating grafana dashboards")
	}
	return nil
}

func (gr *grafana) CleanupOrg(orgName string) error {
	org, err := gr.getOrganization(orgName)
	if err != nil {
		return err
	}
	// org does not exist. exit
	if org == nil {
		return nil
	}

	users, err := gr.usersInOrg(org.Id)
	if err != nil {
		return err
	}
	for _, usr := range users {
		if usr.Role == VIEWER_ROLE || usr.Role == EDITOR_ROLE {
			if err := gr.deleteUser(usr.UserId); err != nil {
				return err
			}
		}
	}
	if err = gr.deleteOrganization(org.Id); err != nil {
		return err
	}
	return nil
}
