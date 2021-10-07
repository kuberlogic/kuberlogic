package util

import (
	"github.com/go-openapi/runtime"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	"net/http"
)

const BadRequestCode int = 400

type BadRequest struct {
	Payload *models.Error `json:"body,omitempty"`
}

// WithPayload adds the payload to the kuberlogic add bad request response
func (o *BadRequest) WithPayload(payload *models.Error) *BadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the kuberlogic add bad request response
func (o *BadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *BadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	rw.WriteHeader(BadRequestCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

func BadRequestFromError(err error) *BadRequest {
	return &BadRequest{
		Payload: &models.Error{
			Message: err.Error(),
		},
	}
}
