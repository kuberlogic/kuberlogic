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

package policy

import (
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/cache"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/logging"
)

type Enforcer interface {
	IsAuthorized(permissions Permissions, user, resource, action string) (bool, error)
}

type PermissionRule struct {
	Subject  string
	Resource string
	Action   string
}

type Permissions struct {
	Rules []PermissionRule
}

func NewEnforcer(cache cache.Cache, log logging.Logger) Enforcer {
	return newCasbinEnforcer(cache, log)
}
