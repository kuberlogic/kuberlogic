// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

// ServiceArchiveOKCode is the HTTP code returned for type ServiceArchiveOK
const ServiceArchiveOKCode int = 200

/*ServiceArchiveOK service request to archive is sent

swagger:response serviceArchiveOK
*/
type ServiceArchiveOK struct {
}

// NewServiceArchiveOK creates ServiceArchiveOK with default headers values
func NewServiceArchiveOK() *ServiceArchiveOK {

	return &ServiceArchiveOK{}
}

// WriteResponse to the client
func (o *ServiceArchiveOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(200)
}

// ServiceArchiveBadRequestCode is the HTTP code returned for type ServiceArchiveBadRequest
const ServiceArchiveBadRequestCode int = 400

/*ServiceArchiveBadRequest invalid input

swagger:response serviceArchiveBadRequest
*/
type ServiceArchiveBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewServiceArchiveBadRequest creates ServiceArchiveBadRequest with default headers values
func NewServiceArchiveBadRequest() *ServiceArchiveBadRequest {

	return &ServiceArchiveBadRequest{}
}

// WithPayload adds the payload to the service archive bad request response
func (o *ServiceArchiveBadRequest) WithPayload(payload *models.Error) *ServiceArchiveBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the service archive bad request response
func (o *ServiceArchiveBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ServiceArchiveBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ServiceArchiveUnauthorizedCode is the HTTP code returned for type ServiceArchiveUnauthorized
const ServiceArchiveUnauthorizedCode int = 401

/*ServiceArchiveUnauthorized bad authentication

swagger:response serviceArchiveUnauthorized
*/
type ServiceArchiveUnauthorized struct {
}

// NewServiceArchiveUnauthorized creates ServiceArchiveUnauthorized with default headers values
func NewServiceArchiveUnauthorized() *ServiceArchiveUnauthorized {

	return &ServiceArchiveUnauthorized{}
}

// WriteResponse to the client
func (o *ServiceArchiveUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// ServiceArchiveForbiddenCode is the HTTP code returned for type ServiceArchiveForbidden
const ServiceArchiveForbiddenCode int = 403

/*ServiceArchiveForbidden bad permissions

swagger:response serviceArchiveForbidden
*/
type ServiceArchiveForbidden struct {
}

// NewServiceArchiveForbidden creates ServiceArchiveForbidden with default headers values
func NewServiceArchiveForbidden() *ServiceArchiveForbidden {

	return &ServiceArchiveForbidden{}
}

// WriteResponse to the client
func (o *ServiceArchiveForbidden) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(403)
}

// ServiceArchiveNotFoundCode is the HTTP code returned for type ServiceArchiveNotFound
const ServiceArchiveNotFoundCode int = 404

/*ServiceArchiveNotFound service not found

swagger:response serviceArchiveNotFound
*/
type ServiceArchiveNotFound struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewServiceArchiveNotFound creates ServiceArchiveNotFound with default headers values
func NewServiceArchiveNotFound() *ServiceArchiveNotFound {

	return &ServiceArchiveNotFound{}
}

// WithPayload adds the payload to the service archive not found response
func (o *ServiceArchiveNotFound) WithPayload(payload *models.Error) *ServiceArchiveNotFound {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the service archive not found response
func (o *ServiceArchiveNotFound) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ServiceArchiveNotFound) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(404)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ServiceArchiveUnprocessableEntityCode is the HTTP code returned for type ServiceArchiveUnprocessableEntity
const ServiceArchiveUnprocessableEntityCode int = 422

/*ServiceArchiveUnprocessableEntity bad validation

swagger:response serviceArchiveUnprocessableEntity
*/
type ServiceArchiveUnprocessableEntity struct {
}

// NewServiceArchiveUnprocessableEntity creates ServiceArchiveUnprocessableEntity with default headers values
func NewServiceArchiveUnprocessableEntity() *ServiceArchiveUnprocessableEntity {

	return &ServiceArchiveUnprocessableEntity{}
}

// WriteResponse to the client
func (o *ServiceArchiveUnprocessableEntity) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(422)
}

// ServiceArchiveServiceUnavailableCode is the HTTP code returned for type ServiceArchiveServiceUnavailable
const ServiceArchiveServiceUnavailableCode int = 503

/*ServiceArchiveServiceUnavailable internal service error

swagger:response serviceArchiveServiceUnavailable
*/
type ServiceArchiveServiceUnavailable struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewServiceArchiveServiceUnavailable creates ServiceArchiveServiceUnavailable with default headers values
func NewServiceArchiveServiceUnavailable() *ServiceArchiveServiceUnavailable {

	return &ServiceArchiveServiceUnavailable{}
}

// WithPayload adds the payload to the service archive service unavailable response
func (o *ServiceArchiveServiceUnavailable) WithPayload(payload *models.Error) *ServiceArchiveServiceUnavailable {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the service archive service unavailable response
func (o *ServiceArchiveServiceUnavailable) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ServiceArchiveServiceUnavailable) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(503)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
