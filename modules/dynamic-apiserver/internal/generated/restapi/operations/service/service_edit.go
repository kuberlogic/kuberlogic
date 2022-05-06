// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// ServiceEditHandlerFunc turns a function with the right signature into a service edit handler
type ServiceEditHandlerFunc func(ServiceEditParams) middleware.Responder

// Handle executing the request and returning a response
func (fn ServiceEditHandlerFunc) Handle(params ServiceEditParams) middleware.Responder {
	return fn(params)
}

// ServiceEditHandler interface for that can handle valid service edit params
type ServiceEditHandler interface {
	Handle(ServiceEditParams) middleware.Responder
}

// NewServiceEdit creates a new http.Handler for the service edit operation
func NewServiceEdit(ctx *middleware.Context, handler ServiceEditHandler) *ServiceEdit {
	return &ServiceEdit{Context: ctx, Handler: handler}
}

/* ServiceEdit swagger:route PUT /services/{ServiceID}/ service serviceEdit

edit a service item

Edit service object


*/
type ServiceEdit struct {
	Context *middleware.Context
	Handler ServiceEditHandler
}

func (o *ServiceEdit) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewServiceEditParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}