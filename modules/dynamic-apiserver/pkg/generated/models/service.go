// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Service service
//
// swagger:model Service
type Service struct {

	// advanced
	Advanced Advanced `json:"advanced,omitempty"`

	// backup schedule
	BackupSchedule string `json:"backupSchedule,omitempty"`

	// created at
	// Read Only: true
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty"`

	// domain
	// Read Only: true
	Domain string `json:"domain,omitempty"`

	// endpoint
	// Read Only: true
	Endpoint string `json:"endpoint,omitempty"`

	// id
	// Required: true
	// Max Length: 20
	// Min Length: 2
	// Pattern: [a-z0-9]([-a-z0-9]*[a-z0-9])?
	ID *string `json:"id"`

	// limits
	Limits *Limits `json:"limits,omitempty"`

	// replicas
	Replicas *int64 `json:"replicas,omitempty"`

	// status
	// Read Only: true
	Status string `json:"status,omitempty"`

	// subscription
	Subscription string `json:"subscription,omitempty"`

	// tls enabled
	TLSEnabled bool `json:"tlsEnabled,omitempty"`

	// type
	// Required: true
	Type *string `json:"type"`

	// version
	Version string `json:"version,omitempty"`
}

// Validate validates this service
func (m *Service) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateAdvanced(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateID(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateLimits(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateType(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Service) validateAdvanced(formats strfmt.Registry) error {
	if swag.IsZero(m.Advanced) { // not required
		return nil
	}

	if m.Advanced != nil {
		if err := m.Advanced.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("advanced")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("advanced")
			}
			return err
		}
	}

	return nil
}

func (m *Service) validateCreatedAt(formats strfmt.Registry) error {
	if swag.IsZero(m.CreatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("created_at", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Service) validateID(formats strfmt.Registry) error {

	if err := validate.Required("id", "body", m.ID); err != nil {
		return err
	}

	if err := validate.MinLength("id", "body", *m.ID, 2); err != nil {
		return err
	}

	if err := validate.MaxLength("id", "body", *m.ID, 20); err != nil {
		return err
	}

	if err := validate.Pattern("id", "body", *m.ID, `[a-z0-9]([-a-z0-9]*[a-z0-9])?`); err != nil {
		return err
	}

	return nil
}

func (m *Service) validateLimits(formats strfmt.Registry) error {
	if swag.IsZero(m.Limits) { // not required
		return nil
	}

	if m.Limits != nil {
		if err := m.Limits.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("limits")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("limits")
			}
			return err
		}
	}

	return nil
}

func (m *Service) validateType(formats strfmt.Registry) error {

	if err := validate.Required("type", "body", m.Type); err != nil {
		return err
	}

	return nil
}

// ContextValidate validate this service based on the context it is used
func (m *Service) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	if err := m.contextValidateAdvanced(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateCreatedAt(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateDomain(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateEndpoint(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateLimits(ctx, formats); err != nil {
		res = append(res, err)
	}

	if err := m.contextValidateStatus(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Service) contextValidateAdvanced(ctx context.Context, formats strfmt.Registry) error {

	if err := m.Advanced.ContextValidate(ctx, formats); err != nil {
		if ve, ok := err.(*errors.Validation); ok {
			return ve.ValidateName("advanced")
		} else if ce, ok := err.(*errors.CompositeError); ok {
			return ce.ValidateName("advanced")
		}
		return err
	}

	return nil
}

func (m *Service) contextValidateCreatedAt(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "created_at", "body", strfmt.DateTime(m.CreatedAt)); err != nil {
		return err
	}

	return nil
}

func (m *Service) contextValidateDomain(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "domain", "body", string(m.Domain)); err != nil {
		return err
	}

	return nil
}

func (m *Service) contextValidateEndpoint(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "endpoint", "body", string(m.Endpoint)); err != nil {
		return err
	}

	return nil
}

func (m *Service) contextValidateLimits(ctx context.Context, formats strfmt.Registry) error {

	if m.Limits != nil {
		if err := m.Limits.ContextValidate(ctx, formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("limits")
			} else if ce, ok := err.(*errors.CompositeError); ok {
				return ce.ValidateName("limits")
			}
			return err
		}
	}

	return nil
}

func (m *Service) contextValidateStatus(ctx context.Context, formats strfmt.Registry) error {

	if err := validate.ReadOnly(ctx, "status", "body", string(m.Status)); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Service) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Service) UnmarshalBinary(b []byte) error {
	var res Service
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
