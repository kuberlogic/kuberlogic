// Code generated by go-swagger; DO NOT EDIT.

package backup

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

// BackupAddCreatedCode is the HTTP code returned for type BackupAddCreated
const BackupAddCreatedCode int = 201

/*BackupAddCreated item created

swagger:response backupAddCreated
*/
type BackupAddCreated struct {

	/*
	  In: Body
	*/
	Payload *models.Backup `json:"body,omitempty"`
}

// NewBackupAddCreated creates BackupAddCreated with default headers values
func NewBackupAddCreated() *BackupAddCreated {

	return &BackupAddCreated{}
}

// WithPayload adds the payload to the backup add created response
func (o *BackupAddCreated) WithPayload(payload *models.Backup) *BackupAddCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the backup add created response
func (o *BackupAddCreated) SetPayload(payload *models.Backup) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *BackupAddCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// BackupAddBadRequestCode is the HTTP code returned for type BackupAddBadRequest
const BackupAddBadRequestCode int = 400

/*BackupAddBadRequest invalid input, object invalid

swagger:response backupAddBadRequest
*/
type BackupAddBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewBackupAddBadRequest creates BackupAddBadRequest with default headers values
func NewBackupAddBadRequest() *BackupAddBadRequest {

	return &BackupAddBadRequest{}
}

// WithPayload adds the payload to the backup add bad request response
func (o *BackupAddBadRequest) WithPayload(payload *models.Error) *BackupAddBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the backup add bad request response
func (o *BackupAddBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *BackupAddBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// BackupAddUnauthorizedCode is the HTTP code returned for type BackupAddUnauthorized
const BackupAddUnauthorizedCode int = 401

/*BackupAddUnauthorized bad authentication

swagger:response backupAddUnauthorized
*/
type BackupAddUnauthorized struct {
}

// NewBackupAddUnauthorized creates BackupAddUnauthorized with default headers values
func NewBackupAddUnauthorized() *BackupAddUnauthorized {

	return &BackupAddUnauthorized{}
}

// WriteResponse to the client
func (o *BackupAddUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(401)
}

// BackupAddConflictCode is the HTTP code returned for type BackupAddConflict
const BackupAddConflictCode int = 409

/*BackupAddConflict item already exists

swagger:response backupAddConflict
*/
type BackupAddConflict struct {
}

// NewBackupAddConflict creates BackupAddConflict with default headers values
func NewBackupAddConflict() *BackupAddConflict {

	return &BackupAddConflict{}
}

// WriteResponse to the client
func (o *BackupAddConflict) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(409)
}

// BackupAddUnprocessableEntityCode is the HTTP code returned for type BackupAddUnprocessableEntity
const BackupAddUnprocessableEntityCode int = 422

/*BackupAddUnprocessableEntity bad validation

swagger:response backupAddUnprocessableEntity
*/
type BackupAddUnprocessableEntity struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewBackupAddUnprocessableEntity creates BackupAddUnprocessableEntity with default headers values
func NewBackupAddUnprocessableEntity() *BackupAddUnprocessableEntity {

	return &BackupAddUnprocessableEntity{}
}

// WithPayload adds the payload to the backup add unprocessable entity response
func (o *BackupAddUnprocessableEntity) WithPayload(payload *models.Error) *BackupAddUnprocessableEntity {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the backup add unprocessable entity response
func (o *BackupAddUnprocessableEntity) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *BackupAddUnprocessableEntity) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(422)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// BackupAddServiceUnavailableCode is the HTTP code returned for type BackupAddServiceUnavailable
const BackupAddServiceUnavailableCode int = 503

/*BackupAddServiceUnavailable internal server error

swagger:response backupAddServiceUnavailable
*/
type BackupAddServiceUnavailable struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewBackupAddServiceUnavailable creates BackupAddServiceUnavailable with default headers values
func NewBackupAddServiceUnavailable() *BackupAddServiceUnavailable {

	return &BackupAddServiceUnavailable{}
}

// WithPayload adds the payload to the backup add service unavailable response
func (o *BackupAddServiceUnavailable) WithPayload(payload *models.Error) *BackupAddServiceUnavailable {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the backup add service unavailable response
func (o *BackupAddServiceUnavailable) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *BackupAddServiceUnavailable) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(503)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}