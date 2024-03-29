// Code generated by go-swagger; DO NOT EDIT.

package service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewServiceListParams creates a new ServiceListParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewServiceListParams() *ServiceListParams {
	return &ServiceListParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewServiceListParamsWithTimeout creates a new ServiceListParams object
// with the ability to set a timeout on a request.
func NewServiceListParamsWithTimeout(timeout time.Duration) *ServiceListParams {
	return &ServiceListParams{
		timeout: timeout,
	}
}

// NewServiceListParamsWithContext creates a new ServiceListParams object
// with the ability to set a context for a request.
func NewServiceListParamsWithContext(ctx context.Context) *ServiceListParams {
	return &ServiceListParams{
		Context: ctx,
	}
}

// NewServiceListParamsWithHTTPClient creates a new ServiceListParams object
// with the ability to set a custom HTTPClient for a request.
func NewServiceListParamsWithHTTPClient(client *http.Client) *ServiceListParams {
	return &ServiceListParams{
		HTTPClient: client,
	}
}

/* ServiceListParams contains all the parameters to send to the API endpoint
   for the service list operation.

   Typically these are written to a http.Request.
*/
type ServiceListParams struct {

	/* SubscriptionID.

	   subscription ID
	*/
	SubscriptionID *string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the service list params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ServiceListParams) WithDefaults() *ServiceListParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the service list params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ServiceListParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the service list params
func (o *ServiceListParams) WithTimeout(timeout time.Duration) *ServiceListParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the service list params
func (o *ServiceListParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the service list params
func (o *ServiceListParams) WithContext(ctx context.Context) *ServiceListParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the service list params
func (o *ServiceListParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the service list params
func (o *ServiceListParams) WithHTTPClient(client *http.Client) *ServiceListParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the service list params
func (o *ServiceListParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithSubscriptionID adds the subscriptionID to the service list params
func (o *ServiceListParams) WithSubscriptionID(subscriptionID *string) *ServiceListParams {
	o.SetSubscriptionID(subscriptionID)
	return o
}

// SetSubscriptionID adds the subscriptionId to the service list params
func (o *ServiceListParams) SetSubscriptionID(subscriptionID *string) {
	o.SubscriptionID = subscriptionID
}

// WriteToRequest writes these params to a swagger request
func (o *ServiceListParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.SubscriptionID != nil {

		// query param SubscriptionID
		var qrSubscriptionID string

		if o.SubscriptionID != nil {
			qrSubscriptionID = *o.SubscriptionID
		}
		qSubscriptionID := qrSubscriptionID
		if qSubscriptionID != "" {

			if err := r.SetQueryParam("SubscriptionID", qSubscriptionID); err != nil {
				return err
			}
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
