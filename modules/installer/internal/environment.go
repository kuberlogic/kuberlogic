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

package internal

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	ImagePullSecret = "kuberlogic-registry"
)

func PrepareEnvironment(namespace string, regServer, regPassword, regUser string, clientSet *kubernetes.Clientset) error {
	ns := &v12.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespace,
		},
	}
	_, err := clientSet.CoreV1().Namespaces().Create(context.TODO(), ns, v1.CreateOptions{})
	if err != nil && !k8serrors.IsAlreadyExists(err) {
		return errors.Wrap(err, "error creating a release namespace")
	}

	if regServer != "" {
		err = createPullSecret(ImagePullSecret, namespace, regServer, regUser, regPassword, clientSet)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			return errors.Wrap(err, "error creating image pull secret")
		}
	}
	return nil
}

func CleanupEnvironment(namespace string, clientSet *kubernetes.Clientset) error {
	if err := deletePullSecret(ImagePullSecret, namespace, true, clientSet); err != nil {
		return errors.Wrap(err, "error deleting image pull secret")
	}
	return nil
}

func createPullSecret(name, ns, server, username, password string, clientset *kubernetes.Clientset) error {
	sec := &v12.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: map[string][]byte{
			".dockerconfigjson": []byte(
				fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}",
					server, username, password, base64.StdEncoding.EncodeToString([]byte(username+":"+password)))),
		},
	}

	// list secrets
	sl, err := clientset.CoreV1().Secrets(ns).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, s := range sl.Items {
		if s.Name == name && s.Namespace == ns {
			return nil // ignore secret that is already exists
		}
	}
	_, err = clientset.CoreV1().Secrets(ns).Create(context.TODO(), sec, v1.CreateOptions{})
	return err
}

func deletePullSecret(name, ns string, force bool, clientset *kubernetes.Clientset) error {
	err := clientset.CoreV1().Secrets(ns).Delete(context.TODO(), name, v1.DeleteOptions{})
	if k8serrors.IsNotFound(err) && force {
		return nil
	}
	return err
}
