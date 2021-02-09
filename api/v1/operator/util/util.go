package util

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"os"
	"strings"
)

const (
	EnvImgRepo       = "IMG_REPO"
	EnvImgPullSecret = "IMG_PULL_SECRET"
)

func GetImage(base, v string) string {
	repo := os.Getenv(EnvImgRepo)
	return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(repo, "/"), base, v)
}

func GetImagePullSecret() string {
	return os.Getenv(EnvImgPullSecret)
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
