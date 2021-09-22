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
	mysqlImage = registryName + "/" + registryOrg + "/mysql-operator:v0.5.0-rc.2-3-gac1ec"

	// monitoring grafana values
	grafanaImageRepo         = registryName + "/" + registryOrg + "/grafana"
	grafanaImageTag          = "8.1.4"
	grafanaServiceName       = "kuberlogic-grafana"
	grafanaServicePort       = 3000
	grafanaSecretName        = "kuberlogic-grafana-credentials"
	grafanaAdminUser         = "kuberlogic"
	grafanaAdminPassword     = "6182ec23cc345656d"
	grafanaMysqlRootPassword = "84fb81edcc1b35"

	// monitoring victoriametrics values
	victoriaMetricsServiceName = "kuberlogic-victoriametrics"

	// operator
	operatorRepository = registryName + "/" + registryOrg

	// apiserver

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
	apiserverTag = "<MUST BE DEFINED>"
	operatorTag  = "<MUST BE DEFINED>"
)
