package grafana

import (
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"net/url"
)

type userCreated struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

type user struct {
	Id    int    `json:"id"`
	OrgId int    `json:"orgId"`
	Email string `json:"email"`
}

type usersOrganization struct {
	OrgId int    `json:"orgId"`
	Name  string `json:"name"`
}

type userOrganization struct {
	OrgId  int    `json:"orgId"`
	UserId int    `json:"userId"`
	Email  string `json:"email"`
	Login  string `json:"login"`
	Role   string `json:"role"`
}

func (gr *grafana) ensureUser(orgId int) error {
	endpoint := "/api/users/lookup"

	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, &url.Values{
		"loginOrEmail": []string{gr.kt.Spec.OwnerEmail},
	})
	if err != nil {
		return err
	}
	switch resp.StatusCode {
	case http.StatusOK:
		gr.log.Info("grafana: user exists", "name", gr.kt.Spec.OwnerEmail)
		var usr user
		if err := gr.api.encodeResponseTo(resp.Body, &usr); err != nil {
			return err
		}
		if err := gr.ensureUserInOrganization(usr, orgId); err != nil {
			return err
		}

		return nil

	case http.StatusNotFound:
		gr.log.Info("grafana: creating a new user", "name", gr.kt.Spec.OwnerEmail)
		return gr.createUser(gr.kt.Spec.OwnerEmail, orgId)

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) ensureUserInOrganization(user user, orgId int) error {
	endpoint := fmt.Sprintf("/api/users/%d/orgs", user.Id)
	resp, err := gr.api.sendRequestTo(http.MethodGet, endpoint, nil)
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
			if item.OrgId == orgId {
				// user is in the organization
				gr.log.Info("grafana: user already in organization",
					"email", user.Email, "orgId", orgId)

				return nil
			}
		}

		return gr.appendUserToOrganization(user, orgId)

	default:
		// something was wrong
		var result interface{}
		err = gr.api.encodeResponseTo(resp.Body, &result)
		return fmt.Errorf("grafana: something was wrong with request %gr, status: %d, result: %v, err: %v",
			endpoint, resp.StatusCode, result, err)
	}
}

func (gr *grafana) createUser(email string, orgId int) error {
	endpoint := "/api/admin/users"
	resp, err := gr.api.sendRequestTo(http.MethodPost,
		endpoint,
		map[string]interface{}{
			"name":     email,
			"email":    email,
			"login":    email,
			"password": uuid.New().String(),
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
	resp, err := gr.api.sendRequestTo(http.MethodDelete, endpoint, nil)
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
