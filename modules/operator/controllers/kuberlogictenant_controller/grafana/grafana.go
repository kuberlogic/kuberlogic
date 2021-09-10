package grafana

import (
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/cfg"
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
