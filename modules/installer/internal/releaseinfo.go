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
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	releaseName       = "kuberlogic"
	releaseSecretName = "kuberlogic-metadata"
	releaseStateKey   = "Status"
	cmSharedSecretKey = "SharedSecret"

	// connection details
	uiURLKey           = "ui"
	mcURLKey           = "mc"
	ingressEndpointKey = "ingressIP"
	demoUserKey        = "demoUser"

	releaseStartedPhase    = "started"
	releaseUpgradePhase    = "upgrading"
	releaseFailedPhase     = "failed"
	releaseSuccessfulPhase = "installed"

	// length of internal password used via Kuberlogic installation process
	internalPasswordLen = 10
)

type ReleaseInfo struct {
	Name      string
	Namespace string
	Status    string

	uiURL           string
	mcURL           string
	ingressEndpoint string
	demoUser        string

	secret *v12.Secret
}

func (r *ReleaseInfo) findState() error {
	if r.secret == nil {
		return errors.New("release secret is not found")
	}
	s, f := r.secret.Data[releaseStateKey]
	if !f || string(s) == "" {
		return errors.New("release state is empty")
	}
	r.Status = string(s)
	r.uiURL = string(r.secret.Data[uiURLKey])
	r.mcURL = string(r.secret.Data[uiURLKey])
	r.ingressEndpoint = string(r.secret.Data[ingressEndpointKey])
	r.demoUser = string(r.secret.Data[demoUserKey])
	return nil
}

func (r *ReleaseInfo) updateState(state string, clientSet *kubernetes.Clientset) error {
	r.secret.Data[releaseStateKey] = []byte(state)
	r.secret.Data[uiURLKey] = []byte(r.uiURL)
	r.secret.Data[mcURLKey] = []byte(r.mcURL)
	r.secret.Data[ingressEndpointKey] = []byte(r.ingressEndpoint)
	r.secret.Data[demoUserKey] = []byte(r.demoUser)

	cm, err := clientSet.CoreV1().Secrets(r.Namespace).Update(context.TODO(), r.secret, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	r.secret = cm
	return nil
}

func (r *ReleaseInfo) updateInfoKey(key, value string, clientSet *kubernetes.Clientset) error {
	r.secret.Data[key] = []byte(value)

	secret, err := clientSet.CoreV1().Secrets(r.Namespace).Update(context.TODO(), r.secret, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	r.secret = secret
	return nil
}

func (r ReleaseInfo) InternalPassword() string {
	return string(r.secret.Data[cmSharedSecretKey])
}

func (r *ReleaseInfo) UpdateUIAddress(addr string) {
	r.uiURL = addr
}

func (r *ReleaseInfo) UpdateMCAddress(addr string) {
	r.mcURL = addr
}

func (r *ReleaseInfo) UpdateIngressAddress(addr string) {
	r.ingressEndpoint = addr
}

func (r *ReleaseInfo) UpdateDemoUser(user string) {
	r.demoUser = user
}

func (r *ReleaseInfo) ShowBanner() string {
	return fmt.Sprintf(`
Kuberlogic URL: %s
Kuberlogic Monitoring Console URL: %s
Kuberlogic address: %s

Demo user login: %s
Demo user password can be found in the configuration file.`, r.uiURL, r.mcURL, r.ingressEndpoint, r.demoUser)
}

func (r *ReleaseInfo) UpgradeRelease(clientSet *kubernetes.Clientset) error {
	err := r.updateState(releaseUpgradePhase, clientSet)
	return err
}

func (r *ReleaseInfo) FinishRelease(clientSet *kubernetes.Clientset) error {
	err := r.updateState(releaseSuccessfulPhase, clientSet)
	return err
}

func StartRelease(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, error) {
	if _, f, err := DiscoverReleaseInfo(namespace, clientSet); err != nil {
		return nil, errors.Wrap(err, "error checking for release")
	} else if f {
		return nil, errors.New("release already exists")
	}
	secret := &v12.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      releaseSecretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			releaseStateKey:   []byte(releaseStartedPhase),
			cmSharedSecretKey: []byte(generateRandomString(internalPasswordLen)),
		},
	}

	s, err := clientSet.CoreV1().Secrets(namespace).Create(context.TODO(), secret, v1.CreateOptions{})
	if err != nil {
		return nil, err
	}

	r := &ReleaseInfo{
		Name:      releaseName,
		Namespace: namespace,
		secret:    s,
	}

	if r.InternalPassword() == "" {
		return nil, errors.New("internal password can't be empty")
	}
	return r, err
}

func FailRelease(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, error) {
	rel, found, err := DiscoverReleaseInfo(namespace, clientSet)
	if !found {
		return nil, errors.New("release is not found")
	}
	if err != nil {
		return nil, errors.Wrap(err, "error finding release")
	}
	err = rel.updateState(releaseFailedPhase, clientSet)
	return rel, err
}

func DiscoverReleaseInfo(namespace string, clientSet *kubernetes.Clientset) (*ReleaseInfo, bool, error) {
	r := &ReleaseInfo{
		Name:      releaseName,
		Namespace: namespace,
	}
	var err error
	r.secret, err = clientSet.CoreV1().Secrets(namespace).Get(context.TODO(), releaseSecretName, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	if r.InternalPassword() == "" {
		return nil, false, errors.New("internal password can't be empty")
	}

	err = r.findState()
	return r, true, nil
}

func CleanupReleaseInfo(namespace string, clientSet *kubernetes.Clientset) error {
	rel, found, err := DiscoverReleaseInfo(namespace, clientSet)
	// release info is not found, exiting
	if !found {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, "error finding release")
	}

	err = clientSet.CoreV1().Secrets(namespace).Delete(context.TODO(), rel.secret.Name, v1.DeleteOptions{})
	return err
}

func generateRandomString(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
