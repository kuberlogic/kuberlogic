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

const VIEWER = "Viewer"

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
		if usr.Role == VIEWER {
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
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, nil)
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
		return nil, fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) createOrganization(orgName string) (int, error) {
	endpoint := "/api/orgs"
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
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
		return 0, fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) deleteOrganization(orgId int) error {
	endpoint := fmt.Sprintf("/api/orgs/%d", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodDelete, endpoint, nil)
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
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) ensureOrganization(orgName string) (int, error) {
	endpoint := fmt.Sprintf("/api/orgs/name/%s", url.QueryEscape(orgName))
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, nil)
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
		return 0, fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) appendUserToOrganization(user user, orgId int) error {
	endpoint := fmt.Sprintf("/api/orgs/%d/users", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
		map[string]interface{}{
			"role":         VIEWER,
			"loginOrEmail": user.Email,
		})
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: user added to organization",
			"name", gr.kt.Spec.OwnerEmail, "email", user.Email, "orgId", orgId)
		return nil

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) usersInOrg(orgId int) ([]*userOrganization, error) {
	endpoint := fmt.Sprintf("/api/orgs/%d/users", orgId)
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, nil)
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
		return nil, fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}
