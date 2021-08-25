package grafana

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/cfg"
)

type grafana struct {
	kt *v1.KuberLogicTenant

	api *API
	log logr.Logger
}

func NewGrafanaSyncer(kt *v1.KuberLogicTenant, log logr.Logger, cfg cfg.Grafana) *grafana {
	return &grafana{
		kt:  kt,
		api: newGrafanaApi(log, cfg),
		log: log,
	}
}

func (gr *grafana) Sync() error {
	// ensure that there is a dedicated Grafana Organization for this tenant
	orgId, err := gr.ensureOrganization(gr.kt.Name)
	if err != nil {
		return fmt.Errorf("error creating grafana organization: %v", err)
	}
	// ensure that there is a user with Viewer role for this organization
	// this user is used by the Kuberlogic tenant user to access dashboards
	if err := gr.ensureUser(gr.kt.Spec.OwnerEmail, "", uuid.New().String(), VIEWER_ROLE, orgId); err != nil {
		return fmt.Errorf("error creating grafana viewer user: %v", err)
	}
	// ensure that there is a user with Editor role for this organization
	// this user is used to create/update tenant dashboards
	editorUsername := gr.kt.Name + "-editor"
	if err := gr.ensureUser("", editorUsername, uuid.New().String(), EDITOR_ROLE, orgId); err != nil {
		return fmt.Errorf("error creating grafana editor user: %v", err)
	}
	if err := gr.ensureDashboards(editorUsername, gr.kt.Status.Services, orgId); err != nil {
		return err
	}
	return nil
}
