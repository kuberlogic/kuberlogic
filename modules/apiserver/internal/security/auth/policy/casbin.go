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

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

const (
	casbinModel = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && globMatch(r.obj, p.obj) && globMatch(r.act, p.act)`
)

type CasbinEnforcer struct {
	model    model.Model
	enforcer *casbin.SyncedEnforcer
	cache    cache.Cache
	log      logging.Logger
}

func newCasbinEnforcer(cache cache.Cache, log logging.Logger) *CasbinEnforcer {
	m, err := model.NewModelFromString(casbinModel)
	if err != nil {
		log.Fatalw("error loading casbin model", "error", err)
	}

	e, err := casbin.NewSyncedEnforcer(m)
	if err != nil {
		log.Fatalw("error creating casbin enforcer", "error", err)
	}

	return &CasbinEnforcer{
		model:    m,
		enforcer: e,
		cache:    cache,
		log:      log,
	}
}

func (c *CasbinEnforcer) IsAuthorized(permissions Permissions, user, resource, action string) (bool, error) {
	c.log.Debugw("checking if user is authorized to do an action",
		"user", user, "resource", resource, "action", action,
		"permissions", permissions)

	for _, p := range permissions.Rules {
		c.enforcer.AddPermissionForUser(user, p.Resource, p.Action)
	}
	defer c.enforcer.DeletePermissionsForUser(user)

	if res, err := c.enforcer.Enforce(user, resource, action); err != nil {
		return false, err
	} else {
		return res, nil
	}
}
