package helm_installer

import "github.com/kuberlogic/operator/modules/installer/internal"

// helm_values.go contains various installation parameters that are later referenced during installation phase
const (
	certManagerNs = "cert-manager"

	// Nginx Ingress Controller
	ingressClass = "kuberlogic-nginx"

	// registry information for installation
	registryName = "quay.io"
	registryOrg  = "kuberlogic"

	// keycloak_ values are passed to both keycloak for configuration and apiserver for consumption
	keycloakClientId     = "apiserver-client"
	keycloakClientSecret = "apiserver-client-secret"
	keycloakRealmName    = "kuberlogic_realm"
	keycloakURL          = "https://keycloak:8443"

	oauthApiserverId = "kuberlogic_apiserver"

	// postgres operator values
	postgresImageRepo      = registryOrg + "/" + "postgres-operator"
	postgresImageTag       = "v1.6.2"
	postgresSecretTemplate = "{username}.{cluster}.credentials"

	// mysql operator values
	mysqlImage = registryName + "/" + registryOrg + "/mysql-operator:v0.5.0-rc.2-1ef93d3"

	// operator
	operatorTag        = "0.0.29"
	operatorRepository = registryName + "/" + registryOrg

	// apiserver
	apiserverTag               = "0.0.29"
	apiserverPort              = 8001
	apiserverDebuglLogsEnabled = "TRUE"
	apiserverAuthProvider      = "keycloak"

	// kuberlogic UI
	uiImageTag = "demo-v8"
)

var (
	globalValues = map[string]interface{}{
		"imagePullSecrets": []map[string]interface{}{
			{"name": internal.ImagePullSecret},
		},
		"monitoringSelector": map[string]interface{}{
			"key":   "core.kuberlogic.com/scrape",
			"value": "yes",
		},
		"monitoringPortAnnotation": "core.kuberlogic.com/scrape-port",
	}
)
