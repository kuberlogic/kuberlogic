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

package helm_installer

import (
	"context"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// secrets are created by Keycloak operator during realm/client provisioning
const (
	realmSecretName = "credential-kuberlogic-realm-kuberlogic-realm-kuberlogic"
)

func waitForKeycloakResources(ns string, clientset *kubernetes.Clientset) error {
	if err := waitForSecretCreation(realmSecretName, ns, clientset); err != nil {
		return errors.Wrap(err, "failed to wait for Keycloak resources")
	}
	return nil
}

func uninstallKuberlogicKeycloak(ns string, force bool, act *action.Configuration, clientset *kubernetes.Clientset, log logger.Logger) error {
	if err := uninstallHelmChart(helmKuberlogicKeycloakCHart, force, act, log); err != nil {
		return err
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted. Waiting for confirmation.")
	if err := waitForSecretDeletion(realmSecretName, ns, clientset); err != nil {
		return errors.Wrap(err, "failed to clean up Keycloak resources")
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted")

	return uninstallHelmChart(helmKeycloakOperatorChart, force, act, log)
}

func waitForSecretCreation(name, ns string, clientset *kubernetes.Clientset) error {
	const waitTimeoutSec = 450

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			continue
		}
		return nil
	}
	return errors.New("secret is not present :" + name)
}

func waitForSecretDeletion(name, ns string, clientset *kubernetes.Clientset) error {
	const waitTimeoutSec = 450

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		continue
	}
	return errors.New("secret is still present: " + name)
}
