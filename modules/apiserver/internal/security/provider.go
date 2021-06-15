package security

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/cache"
	"github.com/kuberlogic/operator/modules/apiserver/internal/config"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/security"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security/auth/provider/keycloak"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security/auth/provider/none"
)

type AuthProviderInterface interface {
	GetAuthenticationSecret(username, password string) (string, error) // returns secret, error
	Authenticate(secret string) (string, string, error)                // returns username, secret, error
	Authorize(username, action, object string) (bool, error)           // return authorization success, error
	CreatePermissionResource(obj string) error
	DeletePermissionResource(obj string) error
}

type AuthProvider struct {
	AuthProviderInterface
}

func (a AuthProvider) GetNamespace(owner string) string {
	n := md5.Sum([]byte(owner))
	return hex.EncodeToString(n[:])
}

func NewAuthProvider(c *config.Config, cache cache.Cache, log logging.Logger) (*AuthProvider, error) {
	var p AuthProviderInterface
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
	return &AuthProvider{
		AuthProviderInterface: p,
	}, nil
}
