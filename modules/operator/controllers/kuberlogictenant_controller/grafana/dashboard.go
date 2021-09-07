package grafana

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	dashdata "github.com/kuberlogic/operator/modules/operator/controllers/kuberlogictenant_controller/grafana/templates"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
	"strings"
	"text/template"
)

type dashboard struct {
	ServiceList []string
	Uid         string
	Title       string
	TenantID    string

	data string
}

func (d dashboard) generateDashboardSpec() (string, error) {
	t, err := template.New("dashboard").Parse(d.data)
	if err != nil {
		return "", errors.Wrap(err, "error creating dashboard template")
	}

	var rendered bytes.Buffer
	if err := t.Execute(&rendered, d); err != nil {
		return "", errors.Wrap(err, "error parsing dashboard template")
	}
	return rendered.String(), nil
}

func newDashboard(orgId int, serviceType string, tenant string, services []string) (*dashboard, error) {
	var tmpl string
	switch serviceType {
	case "mysql":
		tmpl = dashdata.MysqlDashboard
	case "postgresql":
		tmpl = dashdata.PostgresqlDashboard
	default:
		return nil, errors.New("unknown dashboard type")
	}

	uid := md5.Sum([]byte(serviceType + strconv.Itoa(orgId)))
	return &dashboard{
		ServiceList: services,
		Uid:         hex.EncodeToString(uid[:]),
		Title:       serviceType,
		TenantID:    tenant,

		data: strings.ReplaceAll(tmpl, "\n", ""),
	}, nil
}

// listDashboards lists all dashboards for Grafana organization
func (gr *grafana) listDashboards(orgId int) ([]*dashboard, error) {
	endpoint := fmt.Sprintf("/api/search?type=dash-db")
	d := make([]*dashboard, 0)

	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, orgId, nil)
	if err != nil {
		return d, err
	}

	if err := gr.api.encodeResponseTo(resp.Body, d); err != nil {
		return d, err
	}
	return d, nil
}

// ensureDashboard checks if a dashboard has been created in the current int
func (gr *grafana) updateDashboard(d *dashboard, orgId int) error {
	endpoint := "/api/dashboards/db"

	dashboardString, err := d.generateDashboardSpec()
	if err != nil {
		return errors.Wrap(err, "error generating dashboard spec")
	}
	dashboardJson := make(map[string]interface{})
	if err := json.Unmarshal([]byte(dashboardString), &dashboardJson); err != nil {
		fmt.Println(err)
	}

	resp, err := gr.api.sendRequestTo(http.MethodPost, endpoint, orgId, map[string]interface{}{
		"dashboard": dashboardJson,
		"overwrite": true,
	})
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		gr.log.Error(errors.Wrap(err, "error updating Grafana dashboard"), "Grafana API call is failed",
			"dashboard", dashboardString,
			"statusCode", strconv.Itoa(resp.StatusCode),
			"response", fmt.Sprintf("%v", result))
		return errors.Wrap(err, "unexpected Grafana API error")
	}
	return nil
}

// ensureDashboards creates Grafana dashboards under a specific organization for all provisioned service types
func (gr *grafana) ensureDashboards(services map[string]string, orgId int) error {
	// create a map of current services grouped by type
	servicesByType := servicesByType(services)
	for svcType, svcList := range servicesByType {
		d, err := newDashboard(orgId, svcType, gr.kt.GetTenantName(), svcList)
		if err != nil {
			return errors.Wrap(err, "error creating dashboard")
		}

		if err := gr.updateDashboard(d, orgId); err != nil {
			return err
		}
	}
	return nil
}

// servicesByType returns a map of service names by a serviceType key
func servicesByType(svcs map[string]string) map[string][]string {
	res := make(map[string][]string)

	for k, v := range svcs {
		if _, found := res[v]; found {
			res[v] = append(res[v], k)
		} else {
			res[v] = []string{k}
		}
	}
	return res
}
