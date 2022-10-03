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

func (srv *Service) ArchiveKuberlogicService(serviceId *string) {
	ctx := context.Background()
	// Take backup of the service
	archive, err := srv.CreateKuberlogicServiceBackup(ctx, serviceId)
	if err == nil {
		// Wait until backup is done
		archiveID := archive.GetName()
		// maxRetries == 1 2 4 8 16 32 64 128 256 512 1024 2048 4096
		err = srv.WaitForServiceBackup(ctx, serviceId, &archiveID, 13)
		if err != nil {
			msg := "error waiting for service backup"
			srv.log.Errorw(msg, "error", err)
		}

		// Delete all previous backups
		backups, err := srv.ListKuberlogicServiceBackupsByService(ctx, serviceId)
		if err != nil {
			msg := "error listing service backups"
			srv.log.Errorw(msg, "error", err)
		}
		for _, backup := range backups.Items {
			if backup.GetName() == archiveID {
				continue
			}
			err = srv.DeleteKuberlogicServiceBackup(ctx, &backup.Name)
			if err != nil {
				msg := "error deleting service backup"
				srv.log.Errorw(msg, "error", err)
			}
		}
	} else {
		msg := "error taking service backup"
		srv.log.Errorw(msg, "error", err)
	}

	// Delete service
	err = srv.kuberlogicClient.Delete().
		Resource(serviceK8sResource).
		Name(*serviceId).
		Do(ctx).
		Error()
	if err != nil {
		msg := "error deleting service"
		srv.log.Errorw(msg, "error", err)
	}
}
