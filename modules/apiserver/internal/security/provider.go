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

package security

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/cache"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/config"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/generated/security"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/security/auth/provider/keycloak"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/security/auth/provider/none"
)

type AuthProvider interface {
	GetAuthenticationSecret(username, password string) (string, error)          // returns secret, error
	Authenticate(secret string) (string, string, error)                         // returns username, secret, error
	Authorize(principal *models.Principal, action, object string) (bool, error) // return authorization success, error
	CreatePermissionResource(obj string) error
	DeletePermissionResource(obj string) error
}

func NewAuthProvider(c *config.Config, cache cache.Cache, log logging.Logger) (AuthProvider, error) {
	var p AuthProvider
	var e error

	log.Infow("auth provider", "provider", c.Auth.Provider)
	switch c.Auth.Provider {
	case "keycloak":
		p, e = keycloak.NewKeycloakAuthProvider(
			c.Auth.Keycloak.ClientId,
			c.Auth.Keycloak.ClientSecret,
			c.Auth.Keycloak.RealmName,
			c.Auth.Keycloak.Url,
			cache,
			log,
			security.ServiceGrants)
	case "none":
		p, e = none.NewNoneProvider()
	default:
		e = fmt.Errorf("unknown auth provider: " + c.Auth.Provider)
	}

	if e != nil {
		return nil, e
	}
	return p, nil
}
