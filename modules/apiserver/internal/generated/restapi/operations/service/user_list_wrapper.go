// Code generated by go-swagger; DO NOT EDIT.
package service

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging/posthog"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func UserListWrapper(srv operations.Service, next UserListHandlerFunc) (fn UserListHandlerFunc) {
	return func(params UserListParams, principal *models.Principal) middleware.Responder {

		log := srv.GetLogger()

		// check ServiceID param
		ns, name, err := util.SplitID(params.ServiceID)
		if err != nil {
			msg := "incorrect service id"
			log.Errorw(msg, "serviceId", params.ServiceID, "error", err)
			return NewUserListBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		}

		// check auth
		authProvider := srv.GetAuthProvider()
		if authorized, err := authProvider.Authorize(principal.Token, operations.UserListPermission, params.ServiceID); err != nil {
			msg := "auth bad request"
			log.Errorw(msg, "permission", operations.UserListPermission, "serviceId", params.ServiceID, "error", err)
			return NewUserListBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		} else if !authorized {
			log.Errorw("auth forbidden", "permission", operations.UserListPermission, "serviceId", params.ServiceID)
			return NewUserListForbidden()
		}

		// cluster should exists
		err = srv.LookupService(ns, name)
		if err != nil {
			msg := "service does not exist"
			log.Errorw(msg, "error", err)
			return NewUserListBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		}

		// enqueue data to posthog
		posthogMsg := posthog.NewMessage("user-list")
		posthogMsg.With("service-id", params.ServiceID)
		if perr := posthogMsg.Create(); perr != nil {
			msg := "could not enqueue posthog message"
			log.Errorw(msg, "error", perr)
			return NewUserListServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}

		return next(params, principal)
	}
}
