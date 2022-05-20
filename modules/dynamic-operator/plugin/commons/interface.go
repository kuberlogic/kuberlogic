/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PluginService is the interface that we're exposing as a plugin.
type PluginService interface {
	SetLogger(logger hclog.Logger)
	Convert(req PluginRequest) *PluginResponse
	Status(req PluginRequest) *PluginResponseStatus
	Types() *PluginResponse

	Default() *PluginResponseDefault
	ValidateCreate(req PluginRequest) *PluginResponseValidation
	ValidateUpdate(req PluginRequest) *PluginResponseValidation
	ValidateDelete(req PluginRequest) *PluginResponseValidation
}

type PluginRequestEmpty struct{}
type PluginRequest struct {
	Name      string
	Namespace string
	Host      string

	Replicas   int32
	VolumeSize string
	Version    string
	Limits     []byte

	Parameters map[string]interface{}
	Objects    []*unstructured.Unstructured
}

func (pl *PluginRequest) SetLimits(limits *v1.ResourceList) error {
	bytes, _ := json.Marshal(limits)
	pl.Limits = bytes
	return nil
}

func (pl *PluginRequest) GetLimits() (*v1.ResourceList, error) {
	limits := &v1.ResourceList{}
	err := json.Unmarshal(pl.Limits, limits)
	if err != nil {
		return nil, err
	}
	return limits, nil
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
	Objects []*unstructured.Unstructured
	Err     string
}

func (pl *PluginResponse) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}

func (pl *PluginResponse) AddObject(object client.Object, gvk schema.GroupVersionKind) error {
	o, err := ToUnstructured(object, gvk)
	if err != nil {
		pl.Err = err.Error()
		return err
	}
	pl.Objects = append(pl.Objects, o)
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
	Replicas   int32
	VolumeSize string
	Version    string
	Host       string
	// *v1.ResourceList
	Limits     []byte
	Parameters map[string]interface{}
	Err        string
}

func (pl *PluginResponseDefault) SetLimits(limits *v1.ResourceList) error {
	bytes, err := json.Marshal(limits)
	if err != nil {
		log.Fatalf("error when marshaling limits: %v", err)
	}
	pl.Limits = bytes
	return nil
}

func (pl *PluginResponseDefault) GetLimits() *v1.ResourceList {
	if pl.Limits == nil && len(pl.Limits) == 0 {
		return nil
	}

	limits := &v1.ResourceList{}
	err := json.Unmarshal(pl.Limits, limits)
	if err != nil {
		log.Fatalf("error when unmarshaling limits: %v", err)
	}
	return limits
}

func (pl *PluginResponseDefault) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}
