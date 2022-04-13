/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package util

import (
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/models"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
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
	ret.Name = strAsPointer(kls.Name)
	ret.Type = strAsPointer(kls.Spec.Type)
	ret.Replicas = int64AsPointer(int64(kls.Spec.Replicas))
	ret.CreatedAt = strfmt.DateTime(kls.CreationTimestamp.Time.UTC())

	//ret.Resources = kls.Spec.Resources
	//ret.Advanced = kls.Spec.Advanced

	return ret, nil
}

func int64AsPointer(x int64) *int64 {
	return &x
}

func strAsPointer(x string) *string {
	return &x
}

func ErrNotFound(err error) (notFoundErr bool) {
	if statusError, isStatus := err.(*errors.StatusError); isStatus && statusError.Status().Reason == v1.StatusReasonNotFound {
		notFoundErr = true
	}
	return
}
