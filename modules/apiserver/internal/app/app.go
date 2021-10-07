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
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/security"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/store"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Service struct {
	clientset        *kubernetes.Clientset
	kuberlogicClient *rest.RESTClient
	authProvider     security.AuthProvider
	log              logging.Logger
	serviceStore     *store.ServiceStore
}

func (srv *Service) LookupService(ns, name string) (*kuberlogicv1.KuberLogicService, bool, error) {
	item := new(kuberlogicv1.KuberLogicService)
	err := srv.kuberlogicClient.Get().
		Namespace(ns).
		Resource("kuberlogicservices").
		Name(name).
		Do(context.TODO()).
		Into(item)

	if k8s.ErrNotFound(err) {
		return nil, false, err
	} else if err != nil {
		return nil, true, err
	}
	return item, true, nil
}

func (srv *Service) GetLogger() logging.Logger {
	return srv.log
}

func (srv *Service) GetAuthProvider() security.AuthProvider {
	return srv.authProvider
}

func New(clientset *kubernetes.Clientset, client *rest.RESTClient, authProvider security.AuthProvider, log logging.Logger) *Service {
	return &Service{
		clientset:        clientset,
		kuberlogicClient: client,
		authProvider:     authProvider,
		log:              log,
		serviceStore:     store.NewServiceStore(clientset, client, log),
	}
}

func (srv *Service) OnShutdown() {
	defer srv.log.Sync()
}
