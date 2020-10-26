package util

import (
	"fmt"
	"k8s.io/api/core/v1"
	"os"
	"strings"
)

func GetImage(base, v string) string {
	repo := os.Getenv("IMAGE_REPO")
	return fmt.Sprintf("%s/%s:%s", strings.TrimSuffix(repo, "/"), base, v)
}

func GetImagePullSecret() string {
	return os.Getenv("IMAGE_PULL_SECRET")
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
