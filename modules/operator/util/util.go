package util

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/operator/cfg"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

var (
	klRepo, klRepoPullSecret string
)

func InitFromConfig(c *cfg.Config) {
	klRepo = c.ImageRepo
	klRepoPullSecret = c.ImagePullSecretName
}

func GetKuberlogicImage(base, v string) string {
	return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(klRepo, "/"), base, v)
}

func GetKuberlogicRepoPullSecret() string {
	return klRepoPullSecret
}

func FromConfigMap(name, key string) *v1.EnvVarSource {
	return &v1.EnvVarSource{
		ConfigMapKeyRef: &v1.ConfigMapKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: name,
			},
			Key: key,
		},
	}
}

func FromSecret(name, key string) *v1.EnvVarSource {
	return &v1.EnvVarSource{
		SecretKeyRef: &v1.SecretKeySelector{
			LocalObjectReference: v1.LocalObjectReference{
				Name: name,
			},
			Key: key,
		},
	}
}

// converts map[string]string to map[string]intstr.IntOrStr
func StrToIntOrStr(m map[string]string) map[string]intstr.IntOrString {
	var r = make(map[string]intstr.IntOrString, len(m))
	for k, v := range m {
		r[k] = intstr.FromString(v)
	}
	return r
}
