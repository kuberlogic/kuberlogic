/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PluginService is the interface that we're exposing as a plugin.
type PluginService interface {
	SetLogger(logger hclog.Logger)
	Convert(req PluginRequest) *PluginResponse
	Status(req PluginRequest) *PluginResponseStatus
	Type() *PluginResponse

	Default() *PluginResponseDefault
	ValidateCreate(req PluginRequest) *PluginResponseValidation
	ValidateUpdate(req PluginRequest) *PluginResponseValidation
	ValidateDelete(req PluginRequest) *PluginResponseValidation
}

type PluginRequestEmpty struct{}
type PluginRequest struct {
	Name      string
	Namespace string

	Replicas   int32
	VolumeSize string
	Version    string
	Resources  v1.ResourceRequirements

	Parameters map[string]interface{}
	Object     *unstructured.Unstructured
}

type PluginResponseValidation struct {
	Err string
}

func (pl *PluginResponseValidation) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}

type PluginResponse struct {
	Object *unstructured.Unstructured
	Err    string
}

func (pl *PluginResponse) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}

type PluginResponseStatus struct {
	IsReady bool
	Err     string
}

func (pl *PluginResponseStatus) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}

type PluginResponseDefault struct {
	Replicas     int32
	VolumeSize   string
	Version      string
	Resources    v1.ResourceRequirements
	Quantity     resource.Quantity
	ResourceList v1.ResourceList

	Parameters map[string]interface{}

	Err string
}

func (pl *PluginResponseDefault) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}
