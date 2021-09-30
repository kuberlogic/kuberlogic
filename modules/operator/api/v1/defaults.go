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

const MasterUser = "kuberlogic"
const DefaultVolumeSize = "1Gi"

var DefaultResources = v1.ResourceRequirements{
	Requests: v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("10m"),
		v1.ResourceMemory: resource.MustParse("50Mi"),
	},
	Limits: v1.ResourceList{
		// CPU 250m required minimum for zalando/posgtresql
		// Memory 250Mi required minimum for zalando/posgtresql
		v1.ResourceCPU:    resource.MustParse("250m"),
		v1.ResourceMemory: resource.MustParse("512Mi"),
	},
}
