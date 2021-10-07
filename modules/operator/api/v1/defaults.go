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

package v1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type Defaults struct {
	User       string
	VolumeSize string
	Resources  v1.ResourceRequirements
	Version    string
}

const DefaultVolumeSize = "1G"

var DefaultResources = v1.ResourceRequirements{
	Requests: v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("10m"),
		v1.ResourceMemory: resource.MustParse("50M"),
	},
	Limits: v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:    resource.MustParse("250m"),
		v1.ResourceMemory: resource.MustParse("500M"),
	},
}
