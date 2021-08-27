package grafana

import (
	"github.com/go-logr/logr"
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
	orgId, err := gr.ensureOrganization(gr.kt.Name)
	if err != nil {
		return err
	}
	if err := gr.ensureUser(orgId); err != nil {
		return err
	}
	return nil
}
