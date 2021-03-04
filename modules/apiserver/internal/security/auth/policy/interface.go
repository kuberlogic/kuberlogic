package policy

import (
	"github.com/kuberlogic/operator/modules/apiserver/internal/cache"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
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
