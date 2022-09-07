// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

// ServiceListHandlerFunc turns a function with the right signature into a service list handler
type ServiceListHandlerFunc func(ServiceListParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn ServiceListHandlerFunc) Handle(params ServiceListParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// ServiceListHandler interface for that can handle valid service list params
type ServiceListHandler interface {
	Handle(ServiceListParams, *models.Principal) middleware.Responder
}

// NewServiceList creates a new http.Handler for the service list operation
func NewServiceList(ctx *middleware.Context, handler ServiceListHandler) *ServiceList {
	return &ServiceList{Context: ctx, Handler: handler}
}

/* ServiceList swagger:route GET /services/ service serviceList

lists all services

List of service objects


*/
type ServiceList struct {
	Context *middleware.Context
	Handler ServiceListHandler
}

func (o *ServiceList) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewServiceListParams()
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
