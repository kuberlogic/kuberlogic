package grafana

import (
	"fmt"
	"net/http"
	"net/url"
)

type userCreated struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

type user struct {
	Id       int    `json:"id"`
	OrgId    int    `json:"orgId"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

type usersOrganization struct {
	OrgId int    `json:"orgId"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

type userOrganization struct {
	OrgId  int    `json:"orgId"`
	UserId int    `json:"userId"`
	Email  string `json:"email"`
	Login  string `json:"login"`
	Role   string `json:"role"`
}

// ensureUser creates Grafana user if it does not exist
// and appends it to the org identified by orgId
func (gr *grafana) ensureUser(email, username, password, orgRole string, orgId int) error {
	endpoint := "/api/users/lookup"
	if len(email) == 0 {
		email = username
	}
	if len(email) == 0 {
		username = email
	}
	if len(email) == 0 && len(username) == 0 {
		return fmt.Errorf("email or username must be set")
	}

	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, DEFAULT_ORG, &url.Values{
		"loginOrEmail": []string{email},
	})
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: user exists", "username", username, "email", email)
		var usr user
		if err := gr.api.encodeResponseTo(resp.Body, &usr); err != nil {
			return err
		}
		if err := gr.ensureUserInOrganization(usr, orgRole, orgId); err != nil {
			return err
		}

		return nil

	case http.StatusNotFound:
		gr.log.Info("grafana: creating a new user", "name", gr.kt.Spec.OwnerEmail)
		return gr.createUser(email, username, password, orgId)

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

// ensureUserInOrganization checks if a user is added to the organization with id: orgId
// and has orgRole role in this organization
func (gr *grafana) ensureUserInOrganization(user user, orgRole string, orgId int) error {
	endpoint := fmt.Sprintf("/api/users/%d/orgs", user.Id)
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		var result []usersOrganization
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return err
		}
		for _, item := range result {
			if item.OrgId == orgId && item.Role == orgRole {
				// user is in the organization
				gr.log.Info("grafana: user already in organization",
					"email", user.Email, "username", user.Username, "orgId", orgId, "role", orgRole)

				return nil
			}
		}

		return gr.appendUserToOrganization(user, orgRole, orgId)

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

// creates Grafana user
func (gr *grafana) createUser(email, username, password string, orgId int) error {
	endpoint := "/api/admin/users"
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
		DEFAULT_ORG,
		map[string]interface{}{
			"name":     username,
			"email":    email,
			"login":    email,
			"password": password,
			"OrgId":    orgId,
		})
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var result userCreated
		if err := gr.api.encodeResponseTo(resp.Body, &result); err != nil {
			return err
		}
		gr.log.Info("grafana: user is created", "id", result.Id)
		return nil
	default:
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) deleteUser(userId int) error {
	endpoint := fmt.Sprintf("/api/admin/users/%d", userId)
	resp, err := gr.api.sendRequestTo(http.MethodDelete, endpoint, DEFAULT_ORG, nil)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: user is deleted", "id", userId)
		return nil
	default:
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}
