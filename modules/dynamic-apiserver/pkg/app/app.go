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

package app

import (
	"context"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/models"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"time"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

const (
	serviceK8sResource = "kuberlogicservices"
	backupK8sResource  = "kuberlogicservicebackups"
	restoreK8sResource = "kuberlogicservicerestores"
)

type Service struct {
	clientset        kubernetes.Interface
	kuberlogicClient rest.Interface
	log              logging.Logger
	config           *config.Config
}

func (srv *Service) GetLogger() logging.Logger {
	return srv.log
}

func New(cfg *config.Config, clientset kubernetes.Interface, client rest.Interface, log logging.Logger) *Service {
	return &Service{
		clientset:        clientset,
		kuberlogicClient: client,
		log:              log,
		config:           cfg,
	}
}

func (srv *Service) OnShutdown() {
	defer func() {
		_ = srv.log.Sync()
	}()
}

func (srv *Service) ListKuberlogicServiceBackupsByService(ctx context.Context, service *string) (*kuberlogiccomv1alpha1.KuberlogicServiceBackupList, error) {
	res := new(kuberlogiccomv1alpha1.KuberlogicServiceBackupList)

	req := srv.kuberlogicClient.Get().Resource(backupK8sResource)
	if service != nil {
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{util.BackupRestoreServiceField: *service},
		}
		req = req.VersionedParams(&metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}, scheme.ParameterCodec)
	}
	err := req.Do(ctx).Into(res)
	return res, err
}

func (srv *Service) ListKuberlogicServiceRestoresByService(ctx context.Context, service *string) (*kuberlogiccomv1alpha1.KuberlogicServiceRestoreList, error) {
	res := new(kuberlogiccomv1alpha1.KuberlogicServiceRestoreList)

	req := srv.kuberlogicClient.Get().Resource(restoreK8sResource)
	if service != nil {
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{util.BackupRestoreServiceField: *service},
		}
		req = req.VersionedParams(&metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}, scheme.ParameterCodec)
	}
	err := req.Do(ctx).Into(res)
	return res, err
}

func (srv *Service) ListKuberlogicServicesBySubscription(ctx context.Context, subscriptionId *string) (*kuberlogiccomv1alpha1.KuberLogicServiceList, error) {
	res := new(kuberlogiccomv1alpha1.KuberLogicServiceList)

	req := srv.kuberlogicClient.Get().Resource(serviceK8sResource)
	if subscriptionId != nil {
		labelSelector := metav1.LabelSelector{
			MatchLabels: map[string]string{util.SubscriptionField: *subscriptionId},
		}
		req = req.VersionedParams(&metav1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}, scheme.ParameterCodec)
	}
	err := req.Do(ctx).Into(res)
	return res, err
}

func (srv *Service) SubscriptionAlreadyExist(ctx context.Context, subscriptionId *string) (bool, error) {
	services, err := srv.ListKuberlogicServicesBySubscription(ctx, subscriptionId)
	if err != nil {
		return false, err
	}
	return len(services.Items) > 0, nil
}

func (srv *Service) CreateKuberlogicServiceBackup(ctx context.Context, serviceId *string) (*kuberlogiccomv1alpha1.KuberlogicServiceBackup, error) {
	klb := &kuberlogiccomv1alpha1.KuberlogicServiceBackup{
		ObjectMeta: v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%d", *serviceId, time.Now().Unix()),
			Labels: map[string]string{
				util.BackupRestoreServiceField: *serviceId,
			},
		},
		Spec: kuberlogiccomv1alpha1.KuberlogicServiceBackupSpec{
			KuberlogicServiceName: *serviceId,
		},
	}
	err := srv.kuberlogicClient.Post().
		Resource(backupK8sResource).
		Name(klb.GetName()).
		Body(klb).
		Do(ctx).
		Into(klb)
	return klb, err
}

func (srv *Service) DeleteKuberlogicServiceBackup(ctx context.Context, backupId *string) error {
	return srv.kuberlogicClient.Delete().
		Resource(backupK8sResource).
		Name(*backupId).
		Do(ctx).
		Error()
}

func (srv *Service) WaitForServiceBackup(ctx context.Context, serviceId *string, backupId *string, maxRetries int) error {
	timeout := time.Second
	for i := maxRetries; i > 0; i-- {
		backups, err := srv.ListKuberlogicServiceBackupsByService(ctx, serviceId)
		if err != nil {
			return err
		}
		for _, backup := range backups.Items {
			if backup.GetName() == *backupId {
				switch backup.Status.Phase {
				case kuberlogiccomv1alpha1.KlbSuccessfulCondType:
					return nil
				case kuberlogiccomv1alpha1.KlbFailedCondType:
					return errors.New("Error occured while creating service backup")
				}
			}
		}
		timeout = timeout * 2
		time.Sleep(timeout)
		srv.log.Infof("backup is not ready, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, backup is not ready")
}

func (srv *Service) WaitForService(ctx context.Context, serviceId *string, maxRetries int) error {
	timeout := time.Second
	for i := maxRetries; i > 0; i-- {

		result := new(kuberlogiccomv1alpha1.KuberLogicService)
		err := srv.kuberlogicClient.Get().
			Resource(serviceK8sResource).
			Name(*serviceId).
			Do(ctx).
			Into(result)
		if err != nil {
			return errors.Wrap(err, "Error while getting service")
		}

		if ok, _, _ := result.IsReady(); ok {
			return nil
		}

		timeout = timeout * 2
		time.Sleep(timeout)
		srv.log.Infof("service is not ready, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, service is not ready")
}

func (srv *Service) WaitForServiceRestore(ctx context.Context, restoreId string, maxRetries int) error {
	timeout := time.Second
	for i := maxRetries; i > 0; i-- {

		result := new(kuberlogiccomv1alpha1.KuberlogicServiceRestore)
		err := srv.kuberlogicClient.Get().
			Resource(restoreK8sResource).
			Name(restoreId).
			Do(ctx).
			Into(result)
		if err != nil {
			return errors.Wrap(err, "Error while getting service restore")
		}

		if ok := result.IsSuccessful(); ok {
			return nil
		}

		timeout = timeout * 2
		time.Sleep(timeout)
		srv.log.Infof("restore is not successful, trying after %s. Left %d retries", timeout, i-1)
	}
	return errors.New("Retries exceeded, restore is not successful")
}

/*
	Archive service will:
	1. Take new backup of a service (if backups enabled)
	2. Waiting the backup is done
	3. Remove all previous backups
	4. Set "Archive" for the service
*/
func (srv *Service) ArchiveKuberlogicService(service *kuberlogiccomv1alpha1.KuberLogicService) error {
	ctx := context.Background()
	serviceId := &service.Name

	srv.log.Infow("Taking backup of the service", "serviceId", *serviceId)
	archive, err := srv.CreateKuberlogicServiceBackup(ctx, serviceId)
	if err != nil {
		return errors.Wrap(err, "error creating service backup")
	}

	archiveID := archive.GetName()
	maxRetries := 13 // equals 1 2 4 8 16 32 64 128 256 512 1024 2048 4096
	srv.log.Infow("Waiting for backup to be ready", "serviceId", *serviceId, "backupId", archiveID)
	err = srv.WaitForServiceBackup(ctx, serviceId, &archiveID, maxRetries)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	srv.log.Infow("Deleting previous backups", "serviceId", *serviceId)
	backups, err := srv.ListKuberlogicServiceBackupsByService(ctx, serviceId)
	if err != nil {
		return errors.Wrap(err, "error listing service backups")
	}
	for _, backup := range backups.Items {
		if backup.GetName() == archiveID {
			continue
		}
		srv.log.Infow("Deleting backup", "serviceId", *serviceId, "backupId", backup.GetName())
		err = srv.DeleteKuberlogicServiceBackup(ctx, &backup.Name)
		if err != nil {
			return errors.Wrap(err, "error deleting service backup")
		}
	}

	srv.log.Infow("Archive service", "serviceId", *serviceId)
	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err = srv.kuberlogicClient.Patch(types.MergePatchType).
		Resource(serviceK8sResource).
		Name(*serviceId).
		Body([]byte(`{"spec":{"archived":true}}`)). // FIXME: find the better way to do it
		Do(ctx).
		Into(result)
	if err != nil {
		return errors.Wrap(err, "error set archive to service")
	}
	return nil
}

func (srv *Service) FirstSuccessfulBackup(ctx context.Context, serviceId *string) (*kuberlogiccomv1alpha1.KuberlogicServiceBackup, error) {
	backups, err := srv.ListKuberlogicServiceBackupsByService(ctx, serviceId)
	if err != nil {
		return nil, errors.Wrap(err, "error listing service backups")
	}

	for _, backup := range backups.Items {
		switch backup.Status.Phase {
		case kuberlogiccomv1alpha1.KlbSuccessfulCondType:
			return &backup, nil
		}
	}
	return nil, errors.New("No successful backup found")
}

/*
	Unarchive service will:
	1. Find the latest backup
	2. Restoring from the backup
	3. Waiting when the restore completed
	4. Unset "Archive" for the service
*/
func (srv *Service) UnarchiveKuberlogicService(service *kuberlogiccomv1alpha1.KuberLogicService) error {
	ctx := context.Background()
	serviceId := &service.Name

	srv.log.Infow("find successful backup", "serviceId", *serviceId)
	backup, err := srv.FirstSuccessfulBackup(ctx, serviceId)
	if err != nil {
		return errors.Wrap(err, "error finding successful backup")
	}

	klr, err := util.RestoreToKuberlogic(&models.Restore{
		ID:       fmt.Sprintf("%s-%d", backup.GetName(), time.Now().Unix()),
		BackupID: backup.GetName(),
	}, backup)
	if err != nil {
		return errors.Wrap(err, "error creating restore object")
	}

	srv.log.Infow("restore from backup", "serviceId", *serviceId, "backupId", backup.GetName())
	if err := srv.kuberlogicClient.Post().
		Resource(restoreK8sResource).
		Name(klr.GetName()).
		Body(klr).
		Do(ctx).
		Into(klr); k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "restore already exists")
	} else if err != nil {
		return errors.Wrap(err, "error creating restore")
	}

	maxRetries := 13 // equals 1 2 4 8 16 32 64 128 256 512 1024 2048 4096
	srv.log.Infow("Waiting for restore is successful", "serviceId", *serviceId, "restoreId", klr.GetName())
	err = srv.WaitForServiceRestore(ctx, klr.GetName(), maxRetries)
	if err != nil {
		return errors.Wrap(err, "error waiting for service backup")
	}

	srv.log.Infow("unarchive service", "serviceId", *serviceId)
	result := new(kuberlogiccomv1alpha1.KuberLogicService)
	err = srv.kuberlogicClient.Patch(types.MergePatchType).
		Resource(serviceK8sResource).
		Name(*serviceId).
		Body([]byte(`{"spec":{"archived":false}}`)). // FIXME: find the better way to do it
		Do(ctx).
		Into(result)
	if err != nil {
		return errors.Wrap(err, "error set archive to service")
	}
	return nil
}
