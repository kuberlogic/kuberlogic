// Code generated by go-swagger; DO NOT EDIT.

package backup

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
)

// BackupDeleteReader is a Reader for the BackupDelete structure.
type BackupDeleteReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *BackupDeleteReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewBackupDeleteOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewBackupDeleteBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewBackupDeleteUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 403:
		result := NewBackupDeleteForbidden()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 404:
		result := NewBackupDeleteNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 422:
		result := NewBackupDeleteUnprocessableEntity()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 503:
		result := NewBackupDeleteServiceUnavailable()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewBackupDeleteOK creates a BackupDeleteOK with default headers values
func NewBackupDeleteOK() *BackupDeleteOK {
	return &BackupDeleteOK{}
}

/* BackupDeleteOK describes a response with status code 200, with default header values.

item deleted
*/
type BackupDeleteOK struct {
}

func (o *BackupDeleteOK) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteOK ", 200)
}

func (o *BackupDeleteOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewBackupDeleteBadRequest creates a BackupDeleteBadRequest with default headers values
func NewBackupDeleteBadRequest() *BackupDeleteBadRequest {
	return &BackupDeleteBadRequest{}
}

/* BackupDeleteBadRequest describes a response with status code 400, with default header values.

invalid input, object invalid
*/
type BackupDeleteBadRequest struct {
	Payload *models.Error
}

func (o *BackupDeleteBadRequest) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteBadRequest  %+v", 400, o.Payload)
}
func (o *BackupDeleteBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *BackupDeleteBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBackupDeleteUnauthorized creates a BackupDeleteUnauthorized with default headers values
func NewBackupDeleteUnauthorized() *BackupDeleteUnauthorized {
	return &BackupDeleteUnauthorized{}
}

/* BackupDeleteUnauthorized describes a response with status code 401, with default header values.

bad authentication
*/
type BackupDeleteUnauthorized struct {
}

func (o *BackupDeleteUnauthorized) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteUnauthorized ", 401)
}

func (o *BackupDeleteUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewBackupDeleteForbidden creates a BackupDeleteForbidden with default headers values
func NewBackupDeleteForbidden() *BackupDeleteForbidden {
	return &BackupDeleteForbidden{}
}

/* BackupDeleteForbidden describes a response with status code 403, with default header values.

bad permissions
*/
type BackupDeleteForbidden struct {
}

func (o *BackupDeleteForbidden) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteForbidden ", 403)
}

func (o *BackupDeleteForbidden) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewBackupDeleteNotFound creates a BackupDeleteNotFound with default headers values
func NewBackupDeleteNotFound() *BackupDeleteNotFound {
	return &BackupDeleteNotFound{}
}

/* BackupDeleteNotFound describes a response with status code 404, with default header values.

item not found
*/
type BackupDeleteNotFound struct {
}

func (o *BackupDeleteNotFound) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteNotFound ", 404)
}

func (o *BackupDeleteNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewBackupDeleteUnprocessableEntity creates a BackupDeleteUnprocessableEntity with default headers values
func NewBackupDeleteUnprocessableEntity() *BackupDeleteUnprocessableEntity {
	return &BackupDeleteUnprocessableEntity{}
}

/* BackupDeleteUnprocessableEntity describes a response with status code 422, with default header values.

bad validation
*/
type BackupDeleteUnprocessableEntity struct {
}

func (o *BackupDeleteUnprocessableEntity) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteUnprocessableEntity ", 422)
}

func (o *BackupDeleteUnprocessableEntity) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewBackupDeleteServiceUnavailable creates a BackupDeleteServiceUnavailable with default headers values
func NewBackupDeleteServiceUnavailable() *BackupDeleteServiceUnavailable {
	return &BackupDeleteServiceUnavailable{}
}

/* BackupDeleteServiceUnavailable describes a response with status code 503, with default header values.

internal server error
*/
type BackupDeleteServiceUnavailable struct {
	Payload *models.Error
}

func (o *BackupDeleteServiceUnavailable) Error() string {
	return fmt.Sprintf("[DELETE /backups/{BackupID}/][%d] backupDeleteServiceUnavailable  %+v", 503, o.Payload)
}
func (o *BackupDeleteServiceUnavailable) GetPayload() *models.Error {
	return o.Payload
}

func (o *BackupDeleteServiceUnavailable) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}