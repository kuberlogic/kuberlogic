// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

// ServiceGetHandlerFunc turns a function with the right signature into a service get handler
type ServiceGetHandlerFunc func(ServiceGetParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn ServiceGetHandlerFunc) Handle(params ServiceGetParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// ServiceGetHandler interface for that can handle valid service get params
type ServiceGetHandler interface {
	Handle(ServiceGetParams, *models.Principal) middleware.Responder
}

// NewServiceGet creates a new http.Handler for the service get operation
func NewServiceGet(ctx *middleware.Context, handler ServiceGetHandler) *ServiceGet {
	return &ServiceGet{Context: ctx, Handler: handler}
}

/* ServiceGet swagger:route GET /services/{ServiceID}/ service serviceGet

get a service item

Get service object


*/
type ServiceGet struct {
	Context *middleware.Context
	Handler ServiceGetHandler
}

func (o *ServiceGet) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewServiceGetParams()
	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		*r = *aCtx
	}
	var principal *models.Principal
	if uprinc != nil {
		principal = uprinc.(*models.Principal) // this is really a models.Principal, I promise
	}

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params, principal) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
