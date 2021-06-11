// Code generated by go-swagger; DO NOT EDIT.
package service

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/security"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging/posthog"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func UserDeleteWrapper(srv Service, next UserDeleteHandlerFunc) (fn UserDeleteHandlerFunc) {
	return func(params UserDeleteParams, principal *models.Principal) middleware.Responder {

		log := srv.GetLogger()

		// namespace is always provided as a part of Principal object
		ns := principal.Namespace
		// check ServiceID param
		_, name, err := util.SplitID(params.ServiceID)
		if err != nil {
			msg := "incorrect service id"
			log.Errorw(msg, "serviceId", params.ServiceID, "error", err)
			return NewUserDeleteBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		}

		// check auth
		authProvider := srv.GetAuthProvider()
		if authorized, err := authProvider.Authorize(principal.Token, security.UserDeletePermission, params.ServiceID); err != nil {
			msg := "auth bad request"
			log.Errorw(msg, "permission", security.UserDeletePermission, "serviceId", params.ServiceID, "error", err)
			return NewUserDeleteBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		} else if !authorized {
			log.Errorw("auth forbidden", "permission", security.UserDeletePermission, "serviceId", params.ServiceID)
			return NewUserDeleteForbidden()
		}

		// cluster should exists
		service, found, err := srv.LookupService(ns, name)
		if !found {
			msg := "service not found"
			log.Errorw(msg, "error", err)
			return NewUserDeleteBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		} else if err != nil {
			msg := "error getting service"
			log.Errorw(msg, "error", err)
			return NewUserDeleteServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}

		params.HTTPRequest = params.HTTPRequest.WithContext(
			context.WithValue(params.HTTPRequest.Context(), "service", service))

		// enqueue data to posthog
		posthogMsg := posthog.NewMessage("user-delete")
		posthogMsg.With("service-id", params.ServiceID)
		posthogMsg.With("user", params.Username)
		if perr := posthogMsg.Create(); perr != nil {
			msg := "could not enqueue posthog message"
			log.Errorw(msg, "error", perr)
			return NewUserDeleteServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}

		return next(params, principal)
	}
}
