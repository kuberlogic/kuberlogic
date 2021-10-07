/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
