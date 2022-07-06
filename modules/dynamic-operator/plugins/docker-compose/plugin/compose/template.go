/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package compose

import (
	"bytes"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	"html/template"
	"math/rand"
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
	}
}

func (v *viewData) isSecret(value string) bool {
	return strings.Contains(value, "GenerateSecretKey")
}

func (v *viewData) Endpoint(defaultValue string) string {
	var schema, host string
	if v.TLSEnabled {
		schema = "https"
	} else {
		schema = "http"
	}
	if v.Host == "" {
		host = defaultValue
	} else {
		host = v.Host
	}
	return fmt.Sprintf("%s://%s", schema, host)
}

func (v *viewData) parse(value string) (string, error) {
	tmpl, err := template.New("value").Funcs(template.FuncMap{
		"GenerateSecretKey": func(n int) string {
			const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

			b := make([]byte, n)
			for i := range b {
				b[i] = letterBytes[rand.Intn(len(letterBytes))]
			}
			return string(b)
		},
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
