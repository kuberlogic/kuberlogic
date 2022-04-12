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
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/logging"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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
	defer srv.log.Sync()
}
