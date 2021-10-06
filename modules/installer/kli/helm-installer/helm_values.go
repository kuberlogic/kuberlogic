package helm_installer

import "github.com/kuberlogic/kuberlogic/modules/installer/internal"

// helm_values.go contains various installation parameters that are later referenced during installation phase
const (
	certManagerNs = "cert-manager"

	// Kong Ingress Controller
	ingressClass          = "kuberlogic-kong"
	ingressAuthName       = "kuberlogic-auth"
	kongJWTAuthPlugin     = "kuberlogic-jwt-auth"
	kongJWTCleanupPlugin  = "kuberlogic-jwt-param-cleanup"
	kongJWT2HeadersPlugin = "kuberlogic-jwt-headers"
	// registry information for installation
	registryName = "quay.io"
	registryOrg  = "kuberlogic"

	// keycloak_ values are passed to both keycloak for configuration and apiserver for consumption
	keycloakClientId     = "apiserver-client"
	keycloakClientSecret = "apiserver-client-secret"
	keycloakRealmName    = "kuberlogic_realm"
	keycloaInternalkURL  = "https://keycloak:8443"

	keycloakNodePortServiceName = "keycloak-nodeport"

	oauthApiserverId = "kuberlogic_apiserver"

	// jwt parameters used to configure Ingress Controller
	// originated from keycloak
	jwtTokenQueryParam = "token"
	jwtIssuer          = keycloaInternalkURL + "/auth/realms/" + keycloakRealmName

	// postgres operator values
	postgresImageRepo      = registryOrg + "/" + "postgres-operator"
	postgresImageTag       = "v1.6.2"
	postgresSecretTemplate = "{username}.{cluster}.credentials"

	// mysql operator values
	mysqlImage = registryName + "/" + registryOrg + "/mysql-operator:v0.5.0-rc.2-3-gac1ec"

	// monitoring grafana values
	grafanaImageRepo      = registryName + "/" + registryOrg + "/grafana"
	grafanaImageTag       = "8.1.4"
	grafanaServiceName    = "kuberlogic-grafana"
	grafanaServicePort    = 3000
	grafanaSecretName     = "kuberlogic-grafana-credentials"
	grafanaAdminUser      = "kuberlogic"
	grafanaAuthHeaderName = "X-Kong-JWT-Claim-email" // this value is set dynamically by Kong jwt2headers plugin

	// monitoring victoriametrics values
	victoriaMetricsServiceName = "kuberlogic-victoriametrics"

	// operator
	operatorRepository = registryName + "/" + registryOrg

	// apiserver
	apiserverPort              = 8001
	apiserverDebuglLogsEnabled = "TRUE"
	apiserverAuthProvider      = "keycloak"
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
