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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

type API struct {
	query      *url.Values
	baseUrl    string
	username   string
	password   string
	defaultOrg int

	apiMu sync.Mutex
	log   logr.Logger
}

func newHttpClient() *http.Client {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	return &http.Client{Transport: tr}
}

func buildUrl(baseUrl, endpoint string) string {
	endpoint = strings.TrimPrefix(strings.TrimSuffix(endpoint, "/"), "/")
	return fmt.Sprintf("%s%s/", baseUrl, endpoint)
}

func newGrafanaApi(log logr.Logger, cfg cfg.Grafana) *API {
	return &API{
		baseUrl:    cfg.Endpoint,
		username:   cfg.Login,
		password:   cfg.Password,
		defaultOrg: DEFAULT_ORG,
		log:        log,
	}
}

func (api *API) sendRequestTo(method, endpoint string, orgId int, params interface{}) (*http.Response, error) {
	var jsonBody []byte
	var err error
	if params != nil && method != http.MethodGet {
		jsonBody, err = json.Marshal(params)
		if err != nil {
			return nil, err
		}

	}

	req, err := http.NewRequest(method, buildUrl(api.baseUrl, endpoint), bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	if params != nil && method == http.MethodGet {
		req.URL.RawQuery = params.(*url.Values).Encode()
	}

	// handle panic
	defer func() {
		switch t := recover().(type) {
		case string:
			api.log.Error(fmt.Errorf(t), "request to grafana is failed")
		case error:
			api.log.Error(t, "request to grafana is failed")
		}
	}()

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Grafana-Org-Id", strconv.Itoa(orgId))
	req.SetBasicAuth(api.username, api.password)

	client := newHttpClient()
	return client.Do(req)
}

func (api *API) encodeResponseTo(body io.ReadCloser, result interface{}) error {
	content, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(content, &result); err != nil {
		return err
	}
	return nil
}
