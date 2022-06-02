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

package helm_installer

import "github.com/kuberlogic/kuberlogic/modules/installer/internal"

// helm_values.go contains various installation parameters that are later referenced during installation phase
const (
	certManagerNs = "cert-manager"

	// Kong Ingress Controller
	ingressClass                 = "kuberlogic-kong"
	kongJWTCleanupPlugin         = "kuberlogic-jwt-param-cleanup"
	kongKeycloakIntrospectPlugin = "keycloak-introspect-plugin"
	// registry information for installation
	registryName = "quay.io"
	registryOrg  = "kuberlogic"

	// keycloak_ values are passed to both keycloak for configuration and apiserver for consumption
	keycloakDemoUser     = "user@kuberlogic.com"
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

	// monitoring grafana values
	grafanaImageRepo      = registryName + "/" + registryOrg + "/grafana"
	grafanaImageTag       = "8.1.4"
	grafanaServiceName    = "kuberlogic-grafana"
	grafanaServicePort    = 3000
	grafanaSecretName     = "kuberlogic-grafana-credentials"
	grafanaAdminUser      = "kuberlogic"
	grafanaAuthHeaderName = "X-INTROSPECTION-ID"

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
		"kubectlImage":                     "quay.io/bitnami/kubectl:1.21.1",
		"monitoringScrapeAnnotation":       "monitoring.kuberlogic.com/scrape",
		"monitoringScrapeSchemeAnnotation": "monitoring.kuberlogic.com/scheme",
		"monitoringScrapePathAnnotation":   "monitoring.kuberlogic.com/path",
		"monitoringScrapePortAnnotation":   "monitoring.kuberlogic.com/port",
	}
)
