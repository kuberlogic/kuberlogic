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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/logging"
)

type handlers struct {
	clientset  kubernetes.Interface
	restClient rest.Interface
	log        logging.Logger
	config     *config.Config

	services ExtendedServiceInterface
}

var (
	_ Handlers              = &handlers{}
	_ ExtendedServiceGetter = &handlers{}
	_ ExtendedRestoreGetter = &handlers{}
	_ ExtendedBackupGetter  = &handlers{}
)

func (h *handlers) GetLogger() logging.Logger {
	return h.log
}

func New(cfg *config.Config, clientset kubernetes.Interface, client rest.Interface, log logging.Logger) Handlers {
	return &handlers{
		clientset:  clientset,
		restClient: client,
		log:        log,
		config:     cfg,
		services:   newServices(client),
	}
}

func (h *handlers) Services() ExtendedServiceInterface {
	return h.services
}

func (h *handlers) Backups() ExtendedBackupInterface {
	return newBackups(h.restClient)
}

func (h *handlers) Restores() ExtendedRestoreInterface {
	return newRestores(h.restClient)
}

func (h *handlers) OnShutdown() {
	defer func() {
		_ = h.log.Sync()
	}()
}

func (h *handlers) ListOptionsByKeyValue(key string, value *string) v1.ListOptions {
	opts := v1.ListOptions{}
	if value != nil {
		labelSelector := v1.LabelSelector{
			MatchLabels: map[string]string{key: *value},
		}
		opts = v1.ListOptions{
			LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
		}
	}
	return opts
}
