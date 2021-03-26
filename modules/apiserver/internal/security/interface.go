package security

import (
	"fmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/cache"
	"github.com/kuberlogic/operator/modules/apiserver/internal/config"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security/auth/provider/keycloak"
	"github.com/kuberlogic/operator/modules/apiserver/internal/security/auth/provider/none"
)

type AuthProvider interface {
	GetAuthenticationSecret(username, password string) (string, error) // returns secret, error
	Authenticate(secret string) (string, string, error)                // returns username, secret, error
	Authorize(username, action, object string) (bool, error)           // return authorization success, error
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
			log)
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
