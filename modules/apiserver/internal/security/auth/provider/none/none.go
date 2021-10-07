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

package none

import "github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"

const (
	noneEmail  = "none@example.com"
	noneSecret = "secret"
)

type noneAuthProvider struct{}

func (n *noneAuthProvider) GetAuthenticationSecret(username, password string) (string, error) {
	return noneSecret, nil
}

func (n *noneAuthProvider) Authenticate(secret string) (string, string, error) {
	return noneEmail, noneSecret, nil
}

func (n *noneAuthProvider) Authorize(principal *models.Principal, action, object string) (bool, error) {
	return true, nil
}

func (n *noneAuthProvider) CreatePermissionResource(obj string) error {
	return nil
}

func (n *noneAuthProvider) DeletePermissionResource(obj string) error {
	return nil
}

func NewNoneProvider() (*noneAuthProvider, error) {
	return &noneAuthProvider{}, nil
}
