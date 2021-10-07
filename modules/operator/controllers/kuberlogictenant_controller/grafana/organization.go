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
	"fmt"
	"net/http"
	"net/url"
)

type organization struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type organizationCreated struct {
	OrgId   int    `json:"orgId"`
	Message string `json:"message"`
}

const (
	VIEWER_ROLE = "Viewer"
	EDITOR_ROLE = "Editor"
	DEFAULT_ORG = 0
)

func (gr *grafana) DeleteOrganizationAndUsers(orgName string) error {
	org, err := gr.getOrganization(orgName)
	if err != nil {
		return err
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

func (gr *grafana) getOrganization(orgName string) (*organization, error) {
	endpoint := fmt.Sprintf("/api/orgs/name/%s", url.QueryEscape(orgName))
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var result *organization
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return nil, err
		}
		return result, nil

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return nil, fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) createOrganization(orgName string) (int, error) {
	endpoint := "/api/orgs"
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
		DEFAULT_ORG,
		map[string]interface{}{
			"name": orgName,
		})
	if err != nil {
		return 0, err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result organizationCreated
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return 0, err
		}
		gr.log.Info("grafana: organization is created", "name", orgName, "id", result.OrgId)
		return result.OrgId, nil
	default:
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return 0, fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) deleteOrganization(orgId int) error {
	endpoint := fmt.Sprintf("/api/orgs/%d", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodDelete, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: organization deleted", "id", orgId)
		return nil
	default:
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) ensureOrganization(orgName string) (int, error) {
	endpoint := fmt.Sprintf("/api/orgs/name/%s", url.QueryEscape(orgName))
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return 0, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var result organization
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return 0, err
		}
		gr.log.Info("grafana: organization exists", "name", orgName, "id", result.Id)
		return result.Id, nil

	case http.StatusNotFound:
		gr.log.Info("grafana: creating a new organization", "name", orgName)
		return gr.createOrganization(orgName)

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return 0, fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

// appendUserToOrganization adds a user to an organization with identified by orgId with role orgRole
func (gr *grafana) appendUserToOrganization(user user, orgRole string, orgId int) error {
	endpoint := fmt.Sprintf("/api/orgs/%d/users", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
		DEFAULT_ORG,
		map[string]interface{}{
			"role":         orgRole,
			"loginOrEmail": user.Email,
		})
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: user added to organization",
			"name", user.Username, "email", user.Email, "orgId", orgId, "role", orgRole)
		return nil

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) usersInOrg(orgId int) ([]*userOrganization, error) {
	endpoint := fmt.Sprintf("/api/orgs/%d/users", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return nil, err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var result []*userOrganization
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return nil, err
		}
		return result, nil

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return nil, fmt.Errorf("grafana: something was wrong with request %s, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}
