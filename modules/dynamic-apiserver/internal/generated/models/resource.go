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

// Resource resource
//
// swagger:model Resource
type Resource struct {

	// cpu
	// Required: true
	// Pattern: ^([0-9]+$)|([0-9]+.[0-9]+$)
	CPU *string `json:"cpu"`

	// memory
	// Required: true
	// Pattern: ^([0-9]+$)|([0-9]+.[0-9]+$)
	Memory *string `json:"memory"`

	// volume size
	// Required: true
	// Pattern: ^([0-9]+$)|([0-9]+.[0-9]+$)
	VolumeSize *string `json:"volumeSize"`
}

// Validate validates this resource
func (m *Resource) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCPU(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateMemory(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateVolumeSize(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Resource) validateCPU(formats strfmt.Registry) error {

	if err := validate.Required("cpu", "body", m.CPU); err != nil {
		return err
	}

	if err := validate.Pattern("cpu", "body", *m.CPU, `^([0-9]+$)|([0-9]+.[0-9]+$)`); err != nil {
		return err
	}

	return nil
}

func (m *Resource) validateMemory(formats strfmt.Registry) error {

	if err := validate.Required("memory", "body", m.Memory); err != nil {
		return err
	}

	if err := validate.Pattern("memory", "body", *m.Memory, `^([0-9]+$)|([0-9]+.[0-9]+$)`); err != nil {
		return err
	}

	return nil
}

func (m *Resource) validateVolumeSize(formats strfmt.Registry) error {

	if err := validate.Required("volumeSize", "body", m.VolumeSize); err != nil {
		return err
	}

	if err := validate.Pattern("volumeSize", "body", *m.VolumeSize, `^([0-9]+$)|([0-9]+.[0-9]+$)`); err != nil {
		return err
	}

	return nil
}

// ContextValidate validates this resource based on context it is used
func (m *Resource) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Resource) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Resource) UnmarshalBinary(b []byte) error {
	var res Resource
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}