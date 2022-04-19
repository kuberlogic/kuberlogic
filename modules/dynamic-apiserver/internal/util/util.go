/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package util

import (
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ServiceToKuberlogic(svc *models.Service) (*kuberlogiccomv1alpha1.KuberLogicService, error) {
	c := &kuberlogiccomv1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      *svc.Name,
			Namespace: svc.Ns,
		},
	}

	c.Spec.Type = *svc.Type
	if svc.Replicas != nil {
		c.Spec.Replicas = int32(*svc.Replicas)
	}
	if svc.Version != "" {
		c.Spec.Version = svc.Version
	}
	//if svc.Resources != nil {
	//	c.Spec.Resources.Limits = make(v12.ResourceList)
	//	c.Spec.Resources.Requests = make(v12.ResourceList)
	//
	//	cpu := svc.Resources.CPU
	//	if cpu != nil {
	//		// amount of resources and limits could be different
	//		// for using the same values need to use the same defaults in the operator's scope
	//		c.Spec.Resources.Limits[v12.ResourceCPU] = resource.MustParse(*svc.Limits.CPU)
	//	}
	//
	//	mem := svc.Limits.Memory
	//	if mem != nil {
	//		// amount of resources and limits could be different
	//		// for using the same values need to use the same defaults in the operator's scope
	//		c.Spec.Resources.Limits[v12.ResourceMemory] = resource.MustParse(fmt.Sprintf("%vG", *svc.Limits.Memory))
	//	}
	//
	//	if svc.Limits.VolumeSize != nil {
	//		c.Spec.VolumeSize = *svc.Limits.VolumeSize + "G"
	//	}
	//}

	//if svc.Advanced != nil {
	//	c.Spec.Advanced = svc.Advanced
	//}

	return c, nil
}

func KuberlogicToService(kls *kuberlogiccomv1alpha1.KuberLogicService) (*models.Service, error) {
	ret := new(models.Service)
	ret.Name = StrAsPointer(kls.Name)
	ret.Ns = kls.Namespace
	ret.Type = StrAsPointer(kls.Spec.Type)
	ret.Replicas = Int64AsPointer(int64(kls.Spec.Replicas))
	ret.CreatedAt = strfmt.DateTime(kls.CreationTimestamp.Time.UTC())

	if kls.Spec.VolumeSize != "" {
		ret.VolumeSize = kls.Spec.VolumeSize
	}
	if kls.Spec.Limits != nil {
		limits := new(models.Limits)
		if !kls.Spec.Limits.Cpu().IsZero() {
			if value, ok := kls.Spec.Limits[v12.ResourceCPU]; ok {
				limits.CPU = value.String()
			}
		}
		if !kls.Spec.Limits.Memory().IsZero() {
			if value, ok := kls.Spec.Limits[v12.ResourceMemory]; ok {
				limits.Memory = value.String()
			}
		}
		if !kls.Spec.Limits.Storage().IsZero() {
			if value, ok := kls.Spec.Limits[v12.ResourceStorage]; ok {
				limits.VolumeSize = value.String()
			}
		}

		ret.Limits = limits
	}
	if kls.Spec.Version != "" {
		ret.Version = kls.Spec.Version
	}

	_, status := kls.IsReady()
	ret.Status = status

	if kls.Spec.Advanced.Raw != nil {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &ret.Advanced); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func Int64AsPointer(x int64) *int64 {
	return &x
}

func StrAsPointer(x string) *string {
	return &x
}

func ErrNotFound(err error) (notFoundErr bool) {
	if statusError, isStatus := err.(*errors.StatusError); isStatus && statusError.Status().Reason == v1.StatusReasonNotFound {
		notFoundErr = true
	}
	return
}
