package policy

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/cache"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"

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
m = r.sub == p.sub && globMatch(r.act, p.act) && globMatch(r.act, p.act)`
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
