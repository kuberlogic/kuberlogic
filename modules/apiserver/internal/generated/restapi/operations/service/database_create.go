// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"

	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
)

// DatabaseCreateHandlerFunc turns a function with the right signature into a database create handler
type DatabaseCreateHandlerFunc func(DatabaseCreateParams, *models.Principal) middleware.Responder

// Handle executing the request and returning a response
func (fn DatabaseCreateHandlerFunc) Handle(params DatabaseCreateParams, principal *models.Principal) middleware.Responder {
	return fn(params, principal)
}

// DatabaseCreateHandler interface for that can handle valid database create params
type DatabaseCreateHandler interface {
	Handle(DatabaseCreateParams, *models.Principal) middleware.Responder
}

// NewDatabaseCreate creates a new http.Handler for the database create operation
func NewDatabaseCreate(ctx *middleware.Context, handler DatabaseCreateHandler) *DatabaseCreate {
	return &DatabaseCreate{Context: ctx, Handler: handler}
}

/*DatabaseCreate swagger:route POST /services/{ServiceID}/databases/ service databaseCreate

DatabaseCreate database create API

*/
type DatabaseCreate struct {
	Context *middleware.Context
	Handler DatabaseCreateHandler
}

func (o *DatabaseCreate) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewDatabaseCreateParams()

	uprinc, aCtx, err := o.Context.Authorize(r, route)
	if err != nil {
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}
	if aCtx != nil {
		r = aCtx
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
