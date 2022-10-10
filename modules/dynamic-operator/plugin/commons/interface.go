/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
	mathRand "math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
)

type protocol string

const (
	TCPProto  protocol = "tcp"
	HTTPProto protocol = "http"
)

var (
	errSecretNotFound = errors.New("secret not found")
)

// PluginService is the interface that we're exposing as a plugin.
type PluginService interface {
	Convert(req PluginRequest) *PluginResponse
	Status(req PluginRequest) *PluginResponseStatus
	Types() *PluginResponse

	Default() *PluginResponseDefault
	ValidateCreate(req PluginRequest) *PluginResponseValidation
	ValidateUpdate(req PluginRequest) *PluginResponseValidation
	ValidateDelete(req PluginRequest) *PluginResponseValidation

	GetCredentialsMethod(req PluginRequestCredentialsMethod) *PluginResponseCredentialsMethod
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

	// Requested service Version
	Version string

	// If a service should be exposed via TLS
	Insecure bool

	// TLSSecretName is a Kubernetes secret that contains tls.key / tls.crt fields. Must reside in the same namespace
	TLSSecretName string

	// Service resource Limits. Manipulated via SetLimits / GetLimits methods.
	Limits []byte

	// Service will use StorageClass / IngressClass for volume / ingress or default if empty
	StorageClass string
	IngressClass string

	// Additional Parameters
	Parameters map[string]interface{}

	// Credentials
	Credentials map[string]string

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
	b, _ := json.Marshal(limits)
	pl.Limits = b
	return nil
}

func (pl *PluginRequest) GetLimits() (*v1.ResourceList, error) {
	limits := &v1.ResourceList{}
	if pl.Limits != nil && len(pl.Limits) > 0 {
		err := json.Unmarshal(pl.Limits, limits)
		if err != nil {
			return nil, err
		}
	}
	return limits, nil
}

func (pl *PluginRequest) RenderTemplate(tpl string, secrets map[string][]byte) (*TemplatedValue, error) {
	v := &TemplatedValue{}

	tmpl, err := template.New("value").Funcs(template.FuncMap{
		"Secret": func(name string) (string, error) {
			if secret, found := secrets[name]; found {
				v.SecretID = name
				return string(secret), nil
			}
			return "", errSecretNotFound
		},
		"Base64": func(arg string) string {
			return base64.StdEncoding.EncodeToString([]byte(arg))
		},
		"GenerateRSA": func(bits int) (string, error) {
			pk, err := rsa.GenerateKey(rand.Reader, bits)
			if err != nil {
				return "", err
			}

			privateKeyBlock := &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(pk),
			}
			buff := new(bytes.Buffer)
			err = pem.Encode(buff, privateKeyBlock)
			return buff.String(), err
		},
		"GenerateKey": func(length int) string {
			const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

			b := make([]byte, length)
			for i := range b {
				b[i] = letterBytes[mathRand.Intn(len(letterBytes))]
			}
			return string(b)
		},
		"Endpoint": func(defaultValue string) string {
			proto := "https"
			if pl.Insecure {
				proto = "http"
			}

			host := defaultValue
			if pl.Host != "" {
				host = pl.Host
			}
			return fmt.Sprintf("%s://%s", proto, host)
		},
	}).Parse(tpl)
	if err != nil {
		return v, errors.Wrap(err, "failed to parse template data")
	}

	data := &bytes.Buffer{}
	err = tmpl.Execute(data, struct {
		Name       string
		Namespace  string
		Host       string
		Replicas   int32
		Version    string
		Insecure   bool
		Parameters map[string]interface{}
	}{
		Name:       pl.Name,
		Namespace:  pl.Namespace,
		Host:       pl.Host,
		Replicas:   pl.Replicas,
		Version:    pl.Version,
		Insecure:   pl.Insecure,
		Parameters: pl.Parameters,
	})
	v.raw = data.String()
	return v, err
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
	Replicas int32
	Version  string
	Host     string
	// *v1.ResourceList
	Limits     []byte
	Parameters map[string]interface{}
	Err        string
}

func (pl *PluginResponseDefault) SetLimits(limits *v1.ResourceList) error {
	b, err := json.Marshal(limits)
	if err != nil {
		log.Fatalf("error when marshaling limits: %v", err)
	}
	pl.Limits = b
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

type PluginRequestCredentialsMethod struct {
	// service Name
	Name string

	// Credentials Data
	Data map[string]string
}

func (m *PluginRequestCredentialsMethod) RenderTemplate(tpl string) (*TemplatedValue, error) {
	v := &TemplatedValue{}

	tmpl, err := template.New("value").Parse(tpl)
	if err != nil || tmpl == nil {
		return nil, errors.Wrap(err, "failed to parse template")
	}

	data := &bytes.Buffer{}
	err = tmpl.Execute(data, m.Data)
	v.raw = data.String()
	return v, err
}

type PluginResponseCredentialsMethod struct {
	Method string // exec, etc
	Exec   CredentialsMethodExec

	Err string
}

type CredentialsMethodExec struct {
	PodSelector v12.LabelSelector

	Container string
	Command   []string
}

// TemplatedValue is returned on each template value
type TemplatedValue struct {
	raw      string
	SecretID string
}

func (v *TemplatedValue) String() string {
	return v.raw
}
