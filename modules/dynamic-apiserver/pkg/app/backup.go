/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/api"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

type ExtendedBackupGetter interface {
	Backups() ExtendedBackupInterface
}

type ExtendedBackupInterface interface {
	api.BackupInterface
	CreateBackupByServiceName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceBackup, error)
	ListByServiceName(ctx context.Context, service *string) (*v1alpha1.KuberlogicServiceBackupList, error)
	GetEarliestSuccessful(ctx context.Context, serviceId *string) (*v1alpha1.KuberlogicServiceBackup, error)
	Wait(ctx context.Context, log logging.Logger, serviceName *string, backupName string, maxRetries int) error
}
type backups struct {
	api.BackupInterface
}

var _ ExtendedBackupInterface = &backups{}

func newBackups(c rest.Interface) ExtendedBackupInterface {
	s := &backups{}
	s.BackupInterface = api.NewBackups(c)
	return s
}

func (b *backups) CreateBackupByServiceName(ctx context.Context, name string) (*v1alpha1.KuberlogicServiceBackup, error) {
	klb := &v1alpha1.KuberlogicServiceBackup{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", name, time.Now().Unix()),
			Labels: map[string]string{
				util.BackupRestoreServiceField: name,
			},
		},
		Spec: v1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: name,
		},
	}

	return b.Create(ctx, klb, v1.CreateOptions{})
}

func (b *backups) ListByServiceName(ctx context.Context, service *string) (*v1alpha1.KuberlogicServiceBackupList, error) {
	opt := v1.ListOptions{}
	if service != nil {
		labelSelector := v1.LabelSelector{
			MatchLabels: map[string]string{util.BackupRestoreServiceField: *service},
		}
		opt = v1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
	}
	return b.List(ctx, opt)
}

func (b *backups) Wait(ctx context.Context, log logging.Logger, serviceName *string, backupName string, maxRetries int) error {
	timeout := time.Second
	for i := maxRetries; i > 0; i-- {
		r, err := b.ListByServiceName(ctx, serviceName)
		if err != nil {
			return err
		}
		for _, backup := range r.Items {
			if backup.GetName() == backupName {
				switch backup.Status.Phase {
				case v1alpha1.KlbSuccessfulCondType:
					return nil
				case v1alpha1.KlbFailedCondType:
					return errors.New("Error occurred while creating service backup")
				}
			}
		}
		timeout = timeout * 2
		time.Sleep(timeout)
		log.Infof("backup is not ready, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, backup is not ready")
}

func (b *backups) GetEarliestSuccessful(ctx context.Context, serviceId *string) (*v1alpha1.KuberlogicServiceBackup, error) {
	r, err := b.ListByServiceName(ctx, serviceId)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service backups")
	}

	for _, backup := range r.Items {
		switch backup.Status.Phase {
		case v1alpha1.KlbSuccessfulCondType:
			return &backup, nil
		}
	}
	return nil, errors.New("No successful backup found")
}
