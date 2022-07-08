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
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/util"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes/scheme"

	"k8s.io/client-go/kubernetes"
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
}

func (srv *Service) GetLogger() logging.Logger {
	return srv.log
}

func New(clientset kubernetes.Interface, client rest.Interface, log logging.Logger) *Service {
	return &Service{
		clientset:        clientset,
		kuberlogicClient: client,
		log:              log,
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
