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

type protocol string

const (
	TCPproto  protocol = "tcp"
	HTTPproto protocol = "http"
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
	// Requested service Name
	Name string
	// Namespace where the service object must be located
	Namespace string
	// Optional. Host is address by which service should be available.
	Host string

	// Service Replicas
	Replicas int32

	// Requested PV size. VolumeSize should be compatible with Kubernetes ResourceRequirements format
	VolumeSize string

	// Requested service Version
	Version string

	// If a service should be exposed via TLS
	TLSEnabled bool
	// TLSSecretName is a Kubernetes secret that contains tls.key / tls.crt fields. Must reside in the same namespace
	TLSSecretName string

	// Service resource Limits. Manipulated via SetLimits / GetLimits methods.
	Limits []byte

	// Additional Parameters
	Parameters map[string]interface{}

	// Objects contains a list of service related objects
	Objects []*unstructured.Unstructured
}

func (pl *PluginRequest) SetObjects(objs []*unstructured.Unstructured) {
	pl.Objects = objs
}

func (pl *PluginRequest) AddObject(o *unstructured.Unstructured) {
	pl.Objects = append(pl.Objects, o)
}

func (pl *PluginRequest) GetObjects() []*unstructured.Unstructured {
	return pl.Objects
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
	Objects  []*unstructured.Unstructured
	Protocol protocol
	Service  string
	Err      string
}

func (pl *PluginResponse) Error() error {
	if pl.Err != "" {
		return errors.New(pl.Err)
	}
	return nil
}

func (pl *PluginResponse) AddUnstructuredObject(object client.Object, gvk schema.GroupVersionKind) error {
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
