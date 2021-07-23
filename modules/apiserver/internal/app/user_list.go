package app

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	apiService "github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations/service"
	"github.com/kuberlogic/operator/modules/apiserver/internal/store"
	"github.com/kuberlogic/operator/modules/apiserver/util"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	"github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
)

func (srv *Service) UserListHandler(params apiService.UserListParams, principal *models.Principal) middleware.Responder {
	service := params.HTTPRequest.Context().Value("service").(*kuberlogicv1.KuberLogicService)
	session, err := kuberlogic.GetSession(service, srv.clientset, "")
	if err != nil {
		srv.log.Errorw("error generating session", "error", err)
		return util.BadRequestFromError(err)
	}

	users, err := session.GetUser().List()
	if err != nil {
		srv.log.Errorw("error receiving users", "error", err)
		return util.BadRequestFromError(err)
	}

	var payload []*models.User
	for user, permission := range users {
		usr := user
		payload = append(payload, &models.User{
			Name:        &usr,
			Permissions: store.NewPermissionStore().DbToModel(permission),
		})
	}

	return apiService.NewUserListOK().WithPayload(payload)
}
