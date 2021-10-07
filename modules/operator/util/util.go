/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/operator/cfg"
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
