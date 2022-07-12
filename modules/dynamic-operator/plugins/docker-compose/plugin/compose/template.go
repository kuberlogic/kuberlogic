/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package compose

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	"html/template"
	mathRand "math/rand"
	"strings"
)

var (
	errPersistentSecretWrongArgs = errors.New("PersistentSecret function needs `secretId <optional,string>, secretData <string> arguments`")
)

type viewData struct {
	Name       string
	Namespace  string
	Host       string
	Replicas   int32
	VolumeSize string
	Version    string
	TLSEnabled bool
	Limits     []byte
	Parameters map[string]interface{}

	// uses inside template function
	sharedData map[string]string
}

func newViewData(req *commons.PluginRequest) viewData {
	return viewData{
		Name:       req.Name,
		Namespace:  req.Namespace,
		Host:       req.Host,
		Replicas:   req.Replicas,
		VolumeSize: req.VolumeSize,
		Version:    req.Version,
		TLSEnabled: req.TLSEnabled,
		Limits:     req.Limits,
		Parameters: req.Parameters,

		sharedData: make(map[string]string),
	}
}

func (v *viewData) isSecret(value string) bool {
	return strings.Contains(value, "PersistentSecret")
}

func (v *viewData) Endpoint(defaultValue string) string {
	schema := "http"
	if v.TLSEnabled {
		schema = "https"
	}
	host := defaultValue
	if v.Host != "" {
		host = v.Host
	}
	return fmt.Sprintf("%s://%s", schema, host)
}

func (v *viewData) GenerateKey(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[mathRand.Intn(len(letterBytes))]
	}
	return string(b)
}

func (v *viewData) FromCache(key, value string) string {
	if foundValue, ok := v.sharedData[key]; ok {
		return foundValue
	}
	v.sharedData[key] = value
	return value
}

func (v *viewData) GenerateRSA(bits int) (string, error) {
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
}

func (v *viewData) Base64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

// parse executes a template and returns three values:
// * rendered template value
// * value identified (set only when PersistentSecret function is set)
// *
func (v *viewData) parse(value string) (string, string, error) {
	var keyId string
	tmpl, err := template.New("value").Funcs(template.FuncMap{
		// PersistentSecret func accepts two args (one is optional):
		// * keyId (this will be used to identify a secret key), empty if not passed
		// * data (data that needs to be stored in secret)
		"PersistentSecret": func(args ...string) (string, error) {
			switch len(args) {
			case 1:
				return args[0], nil
			case 2:
				keyId = args[0]
				return args[1], nil
			default:
				return "", errPersistentSecretWrongArgs
			}
		},
	}).Parse(value)
	if err != nil {
		return "", "", errors.Wrap(err, "error parsing template")
	}

	data := &bytes.Buffer{}
	if err := tmpl.Execute(data, v); err != nil {
		return "", "", errors.Wrap(err, "error rendering value")
	}
	return data.String(), keyId, nil
}
