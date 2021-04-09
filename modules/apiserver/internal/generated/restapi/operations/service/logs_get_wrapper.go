// Code generated by go-swagger; DO NOT EDIT.
package service

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/restapi/operations"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging/posthog"
	"github.com/kuberlogic/operator/modules/apiserver/util"
)

func LogsGetWrapper(srv operations.Service, next LogsGetHandlerFunc) (fn LogsGetHandlerFunc) {
	return func(params LogsGetParams, principal *models.Principal) middleware.Responder {

		log := srv.GetLogger()

		// check ServiceID param
		ns, name, err := util.SplitID(params.ServiceID)
		if err != nil {
			msg := "incorrect service id"
			log.Errorw(msg, "serviceId", params.ServiceID, "error", err)
			return NewLogsGetBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		}

		// check auth
		authProvider := srv.GetAuthProvider()
		if authorized, err := authProvider.Authorize(principal.Token, operations.LogsGetPermission, params.ServiceID); err != nil {
			msg := "auth bad request"
			log.Errorw(msg, "permission", operations.LogsGetPermission, "serviceId", params.ServiceID, "error", err)
			return NewLogsGetBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		} else if !authorized {
			log.Errorw("auth forbidden", "permission", operations.LogsGetPermission, "serviceId", params.ServiceID)
			return NewLogsGetForbidden()
		}

		// cluster should exists
		err = srv.LookupService(ns, name)
		if err != nil {
			msg := "service does not exist"
			log.Errorw(msg, "error", err)
			return NewLogsGetBadRequest().WithPayload(&models.Error{
				Message: msg,
			})
		}

		// enqueue data to posthog
		posthogMsg := posthog.NewMessage("logs-get")
		posthogMsg.With("service-id", params.ServiceID)
		posthogMsg.With("service-instance", params.ServiceInstance)
		posthogMsg.With("tail", params.Tail)
		if perr := posthogMsg.Create(); perr != nil {
			msg := "could not enqueue posthog message"
			log.Errorw(msg, "error", perr)
			return NewLogsGetServiceUnavailable().WithPayload(&models.Error{
				Message: msg,
			})
		}

		return next(params, principal)
	}
}
