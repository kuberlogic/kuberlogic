package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
)

func (srv *Service) UserListHandler(params apiService.UserListParams, principal *models.Principal) middleware.Responder {
	session, err := kuberlogic.GetSession(srv.existingService, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	users, err := session.GetUser().List()
	if err != nil {
		srv.log.Errorw("error receiving databases", "error", err)
		return util.BadRequestFromError(err)
	}

	var payload []*models.User
	for _, dbUser := range users {
		userName := dbUser

		if protected := session.GetUser().IsProtected(userName); !protected {
			payload = append(payload, &models.User{
				Name: &userName,
			})
		}
	}

	return apiService.NewUserListOK().WithPayload(payload)
}
