/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package util

import (
	"encoding/json"
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	errors2 "github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	SubscriptionField         = "subscription-id"
	BackupRestoreServiceField = "kls-id"
)

func ServiceToKuberlogic(svc *models.Service) (*kuberlogiccomv1alpha1.KuberLogicService, error) {
	c := &kuberlogiccomv1alpha1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name: *svc.ID,
		},
	}

	c.Spec.Type = *svc.Type
	if svc.Replicas != nil {
		c.Spec.Replicas = int32(*svc.Replicas)
	}
	if svc.Version != "" {
		c.Spec.Version = svc.Version
	}
	if svc.Domain != "" {
		c.Spec.Domain = svc.Domain
	}

	if svc.BackupSchedule != "" {
		c.Spec.BackupSchedule = svc.BackupSchedule
	}

	if svc.Limits != nil {
		c.Spec.Limits = make(v12.ResourceList)

		if svc.Limits.CPU != "" {
			// amount of resources and limits could be different
			// for using the same values need to use the same defaults in the operator's scope
			c.Spec.Limits[v12.ResourceCPU] = resource.MustParse(svc.Limits.CPU)
		}

		if svc.Limits.Memory != "" {
			// amount of resources and limits could be different
			// for using the same values need to use the same defaults in the operator's scope
			c.Spec.Limits[v12.ResourceMemory] = resource.MustParse(svc.Limits.Memory)
		}

		if svc.Limits.VolumeSize != "" {
			c.Spec.Limits[v12.ResourceStorage] = resource.MustParse(svc.Limits.VolumeSize)
		}
	}
	c.Spec.TLSEnabled = svc.TLSEnabled

	if svc.Advanced != nil {
		data, err := json.Marshal(svc.Advanced)
		if err != nil {
			return nil, errors2.Wrap(err, "cannot deserialize advanced parameter")
		}
		c.Spec.Advanced.Raw = data
	}

	if svc.Subscription != "" {
		if c.Labels == nil {
			c.Labels = make(map[string]string)
		}
		c.Labels[SubscriptionField] = svc.Subscription
	}

	return c, nil
}

func KuberlogicToService(kls *kuberlogiccomv1alpha1.KuberLogicService) (*models.Service, error) {
	ret := new(models.Service)
	ret.ID = StrAsPointer(kls.Name)
	ret.Type = StrAsPointer(kls.Spec.Type)
	ret.Replicas = Int64AsPointer(int64(kls.Spec.Replicas))
	ret.CreatedAt = strfmt.DateTime(kls.CreationTimestamp.Time.UTC())

	if kls.Spec.Domain != "" {
		ret.Domain = kls.Spec.Domain
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

	if kls.Spec.BackupSchedule != "" {
		ret.BackupSchedule = kls.Spec.BackupSchedule
	}

	ret.Status = kls.Status.Phase
	ret.Endpoint = kls.Status.AccessEndpoint

	if kls.Spec.Advanced.Raw != nil {
		if err := json.Unmarshal(kls.Spec.Advanced.Raw, &ret.Advanced); err != nil {
			return nil, err
		}
	}
	if kls.ObjectMeta.Labels != nil {
		if value, ok := kls.ObjectMeta.Labels["subscription-id"]; ok {
			ret.Subscription = value
		}
	}

	return ret, nil
}

func BackupToKuberlogic(backup *models.Backup) (*kuberlogiccomv1alpha1.KuberlogicServiceBackup, error) {
	return &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: v1.ObjectMeta{
			Name: backup.ID,
			Labels: map[string]string{
				BackupRestoreServiceField: backup.ServiceID,
			},
		},
		Spec: kuberlogiccomv1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: backup.ServiceID,
		},
	}, nil
}

func KuberlogicToBackup(backup *kuberlogiccomv1alpha1.KuberlogicServiceBackup) (*models.Backup, error) {
	return &models.Backup{
		CreatedAt: strfmt.DateTime(backup.GetCreationTimestamp().Time),
		ID:        backup.GetName(),
		ServiceID: backup.Spec.KuberlogicServiceName,
		Status:    backup.Status.Phase,
	}, nil
}

func RestoreToKuberlogic(restore *models.Restore, klb *kuberlogiccomv1alpha1.KuberlogicServiceBackup) (*kuberlogiccomv1alpha1.KuberlogicServiceRestore, error) {
	return &kuberlogiccomv1alpha1.KuberlogicServiceRestore{
		ObjectMeta: v1.ObjectMeta{
			Name: restore.ID,
			Labels: map[string]string{
				BackupRestoreServiceField: klb.Spec.KuberlogicServiceName,
			},
		},
		Spec: kuberlogiccomv1alpha1.KuberlogicServiceRestoreSpec{
			KuberlogicServiceBackup: restore.BackupID,
		},
	}, nil
}

func KuberlogicToRestore(restore *kuberlogiccomv1alpha1.KuberlogicServiceRestore) (*models.Restore, error) {
	return &models.Restore{
		BackupID:  restore.Spec.KuberlogicServiceBackup,
		ID:        restore.GetName(),
		Status:    restore.Status.Phase,
		CreatedAt: strfmt.DateTime(restore.GetCreationTimestamp().Time),
	}, nil
}

func Int64AsPointer(x int64) *int64 {
	return &x
}

func StrAsPointer(x string) *string {
	return &x
}

func CheckStatus(err error, reason v1.StatusReason) bool {
	statusError, ok := err.(*errors.StatusError)
	return ok && statusError.Status().Reason == reason
}
