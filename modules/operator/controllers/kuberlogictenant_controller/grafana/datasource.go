package grafana

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"io"
	"net/http"
)

type datasource struct {
	Name      string `json:"name"`
	Id        int    `json:"id,omitempty"`
	Uid       string `json:"uid,omitempty"`
	Type      string `json:"type"`
	Url       string `json:"url"`
	Access    string `json:"access"`
	BasicAuth bool   `json:"basicAuth,omitempty"`
}

const (
	dsName   = "kuberlogic-monitoring"
	dsType   = "prometheus"
	dsAccess = "proxy"
	dsAuth   = false

	dsApiEndpoint = "/api/datasources"
)

func newDatasource(uid, addr string) *datasource {
	return &datasource{
		Name:      dsName,
		Uid:       uid,
		Type:      dsType,
		Url:       addr,
		Access:    dsAccess,
		BasicAuth: dsAuth,
	}
}

func (gr *grafana) ensureDatasource(orgId int) error {
	uid := gr.kt.Name

	d := newDatasource(uid, gr.datasourceAddress)

	ds, err := gr.getDatasource(uid, orgId)
	if err != nil {
		return errors.Wrap(err, "error getting Grafana datasource")
	}

	if ds != nil {
		if err := gr.updateDatasource(ds, d, orgId); err != nil {
			return errors.Wrap(err, "error updating Grafana datasource")
		}
		return nil
	}

	gr.log.Info("datasource is not found, creating one")
	if gr.createDatasource(d, orgId); err != nil {
		return errors.Wrap(err, "error creating Grafana datasource")
	}
	return nil
}

func (gr *grafana) createDatasource(d *datasource, orgId int) error {
	data := make(map[string]interface{})
	if err := mapstructure.Decode(d, &data); err != nil {
		return errors.Wrap(err, "error decoding request values")
	}
	delete(data, "Id")
	resp, err := gr.api.sendRequestTo(http.MethodPost, dsApiEndpoint, orgId, data)

	if err != nil || resp.StatusCode != http.StatusOK {
		gr.log.Error(err, "error creating Grafana datasource", "code", resp.StatusCode, "data", fmt.Sprintf("%v", data))
		return errors.Wrap(err, "error creating Grafana datasource")
	}
	return nil
}

// updataDatasource checks if current datasource matches desired and updates it in Grafana if needed
func (gr *grafana) updateDatasource(current, desired *datasource, orgId int) error {
	if current.Name == desired.Name &&
		current.Type == desired.Type &&
		current.Url == desired.Url &&
		current.BasicAuth == desired.BasicAuth &&
		current.Access == desired.Access {
		gr.log.Info("Grafana datasource is up to date")
		return nil
	}

	desired.Type = current.Type
	desired.Url = current.Url
	desired.BasicAuth = current.BasicAuth
	desired.Access = current.Access

	data := make(map[string]interface{})
	if err := mapstructure.Decode(desired, &data); err != nil {
		return errors.Wrap(err, "error decoding request data")
	}
	resp, err := gr.api.sendRequestTo(http.MethodPut, fmt.Sprintf("%s/%d", dsApiEndpoint, desired.Id), orgId, data)
	if err != nil {
		gr.log.Error(err, "error sending update datasource request")
		return err
	}
	if resp.StatusCode != http.StatusOK {
		nonOkErr := errors.New("unexpected http status code")
		gr.log.Error(nonOkErr, "HTTP status code is not 200",
			"code", resp.StatusCode, "data", fmt.Sprintf("%v", data))
		return nonOkErr
	}
	return nil
}

// getDatasource gets datasource by name
func (gr *grafana) getDatasource(name string, orgId int) (*datasource, error) {
	resp, err := gr.api.sendRequestTo(http.MethodGet, fmt.Sprintf("%s/name/%s", dsApiEndpoint, name), orgId, nil)
	if err != nil {
		gr.log.Error(err, "error sending get request to Grafana", "code", resp.StatusCode)
		return nil, err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, nil
	case http.StatusOK:
		ds := new(datasource)
		if gr.api.encodeResponseTo(resp.Body, ds); err != nil {
			gr.log.Error(err, "error decoding Grafana response")
			return nil, err
		}
		return ds, nil
	default:
		respData, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding Grafana response")
		}
		nonOkCodeErr := errors.New("incorrect HTTP status code")
		gr.log.Error(nonOkCodeErr, "unexpected status code",
			"code", resp.StatusCode, "response", string(respData))
		return nil, nonOkCodeErr
	}
}
