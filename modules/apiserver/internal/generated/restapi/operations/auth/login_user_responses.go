// Code generated by go-swagger; DO NOT EDIT.

package auth

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
)

// LoginUserOKCode is the HTTP code returned for type LoginUserOK
const LoginUserOKCode int = 200

/*LoginUserOK access token response

swagger:response loginUserOK
*/
type LoginUserOK struct {

	/*
	  In: Body
	*/
	Payload *models.AccessTokenResponse `json:"body,omitempty"`
}

// NewLoginUserOK creates LoginUserOK with default headers values
func NewLoginUserOK() *LoginUserOK {

	return &LoginUserOK{}
}

// WithPayload adds the payload to the login user o k response
func (o *LoginUserOK) WithPayload(payload *models.AccessTokenResponse) *LoginUserOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the login user o k response
func (o *LoginUserOK) SetPayload(payload *models.AccessTokenResponse) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *LoginUserOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// LoginUserBadRequestCode is the HTTP code returned for type LoginUserBadRequest
const LoginUserBadRequestCode int = 400

/*LoginUserBadRequest bad input parameters

swagger:response loginUserBadRequest
*/
type LoginUserBadRequest struct {
}

// NewLoginUserBadRequest creates LoginUserBadRequest with default headers values
func NewLoginUserBadRequest() *LoginUserBadRequest {

	return &LoginUserBadRequest{}
}

// WriteResponse to the client
func (o *LoginUserBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(400)
}

// LoginUserUnauthorizedCode is the HTTP code returned for type LoginUserUnauthorized
const LoginUserUnauthorizedCode int = 401

/*LoginUserUnauthorized bad authentication

swagger:response loginUserUnauthorized
*/
type LoginUserUnauthorized struct {
}

// NewLoginUserUnauthorized creates LoginUserUnauthorized with default headers values
func NewLoginUserUnauthorized() *LoginUserUnauthorized {

	return &LoginUserUnauthorized{}
}

// WriteResponse to the client
func (o *LoginUserUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// LoginUserServiceUnavailableCode is the HTTP code returned for type LoginUserServiceUnavailable
const LoginUserServiceUnavailableCode int = 503

/*LoginUserServiceUnavailable internal server error

swagger:response loginUserServiceUnavailable
*/
type LoginUserServiceUnavailable struct {
}

// NewLoginUserServiceUnavailable creates LoginUserServiceUnavailable with default headers values
func NewLoginUserServiceUnavailable() *LoginUserServiceUnavailable {

	return &LoginUserServiceUnavailable{}
}

// WriteResponse to the client
func (o *LoginUserServiceUnavailable) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(503)
}
