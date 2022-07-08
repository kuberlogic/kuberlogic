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
	return strings.Contains(value, "| KeepSecret")
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

func (v *viewData) GenerateRSA(bits int) string {
	// TODO: how to return err?
	pk, _ := rsa.GenerateKey(rand.Reader, bits)

	privateKeyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(pk),
	}
	buff := new(bytes.Buffer)
	_ = pem.Encode(buff, privateKeyBlock)
	return buff.String()
}

func (v *viewData) Base64(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func (v *viewData) parse(value string) (string, error) {
	tmpl, err := template.New("value").Funcs(template.FuncMap{
		"KeepSecret": func(s string) string { return s }, // just do nothing, flag to store key/value in k8s secrets
	}).Parse(value)
	if err != nil {
		return "", errors.Wrap(err, "error parsing template")
	}

	data := &bytes.Buffer{}
	if err := tmpl.Execute(data, v); err != nil {
		return "", errors.Wrap(err, "error rendering value")
	}
	return data.String(), nil
}
