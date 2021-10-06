package helm_installer

import (
	"context"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/installer/internal"
	"github.com/pkg/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func deployCRDs(globals map[string]interface{}, i *HelmInstaller) error {
	values := make(map[string]interface{}, 0)

	chart, err := crdsChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading CRDs chart")
	}

	i.Log.Infof("Deploying CRDs...")
	return releaseHelmChart(helmCRDsChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployCertManager(globals map[string]interface{}, i *HelmInstaller) error {
	values := map[string]interface{}{
		"installCRDs": true,
	}

	chart, err := certManagerChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading cert-manager chart")
	}

	i.Log.Infof("Deploying cert-manager...")
	return releaseHelmChart(helmCertManagerChart, certManagerNs, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployAuth(globals map[string]interface{}, i *HelmInstaller) error {
	keycloakLocalValues := map[string]interface{}{
		"installCRDs": false,
	}

	operatorChart, err := keycloakOperatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading keycloak-operator chart")
	}

	i.Log.Infof("Deploying keycloak-operator...")
	if err := releaseHelmChart(helmKeycloakOperatorChart, i.ReleaseNamespace, operatorChart, keycloakLocalValues, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying keycloak-operator")
	}

	kuberlogicKeycloakValues := map[string]interface{}{
		"realm": map[string]interface{}{
			"id":            keycloakRealmName,
			"name":          keycloakRealmName,
			"adminPassword": i.Auth.AdminPassword,
		},
		"clientId":     keycloakClientId,
		"clientSecret": keycloakClientSecret,

		"apiserverId": oauthApiserverId,

		"nodePortService": map[string]interface{}{
			"name": keycloakNodePortServiceName,
		},
	}
	if i.Auth.TestUserPassword != "" {
		kuberlogicKeycloakValues["testUser"] = map[string]interface{}{
			"create":   true,
			"password": i.Auth.TestUserPassword,
		}
	}

	kuberlogicKeycloakChart, err := kuberlogicKeycloakChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading kuberlogic-keycloak chart")
	}
	i.Log.Infof("Deploying Kuberlogic Keycloak resources...")
	if err := releaseHelmChart(helmKuberlogicKeycloakCHart, i.ReleaseNamespace, kuberlogicKeycloakChart, kuberlogicKeycloakValues, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying kuberlogic-keycloak")
	}
	if err := waitForKeycloakResources(i.ReleaseNamespace, i.ClientSet); err != nil {
		return errors.Wrap(err, "keycloak provisioning error")
	}
	return nil
}

func deployIngressController(globals map[string]interface{}, i *HelmInstaller, releaseInfo *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"ingressController": map[string]interface{}{
			"installCRDs":  false,
			"ingressClass": ingressClass,
		},
	}
	chart, err := kongIngressControllerChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading kong ingress controller chart")
	}
	i.Log.Infof("Deploying Kong Ingress Controller...")
	if err := releaseHelmChart(helmKongIngressControllerChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying kong ingress controller")
	}

	// verify that Kong Ingress Controller services is created and received Ingress IP address
	// service name is expected to be always the same
	const ingressSvcName = "kong-kong-proxy"
	const waitTimeoutSec = 30
	foundIP := false
	for x := 0; x < waitTimeoutSec; x += 1 {
		time.Sleep(time.Second)
		s, err := i.ClientSet.CoreV1().Services(i.ReleaseNamespace).Get(context.TODO(), ingressSvcName, v1.GetOptions{})
		if err != nil {
			continue // hope that the error is transient
		}
		if len(s.Status.LoadBalancer.Ingress) != 0 {
			// success. append to the release banner
			releaseInfo.UpdateIngressAddress(s.Status.LoadBalancer.Ingress[0].IP)
			foundIP = true
			break
		}
	}
	if !foundIP {
		return errors.New("failed to obtain an Ingress IP address for Kong ingress controller")
	}

	// get authentication data for Kong Ingress Controller
	JWTAuthParams, err := getJWTAuthVals(i.ReleaseNamespace, i.ClientSet, i.Log)
	if err != nil {
		return errors.Wrap(err, "error computing Grafana Authentication values")
	}
	kuberlogicIngressValues := map[string]interface{}{
		"kong": map[string]interface{}{
			"authPlugin":         kongJWTAuthPlugin,
			"tokenCleanupPlugin": kongJWTCleanupPlugin,
			"jwt2headerPlugin": map[string]interface{}{
				"name": kongJWT2HeadersPlugin,
			},
		},
		"jwtAuth":      JWTAuthParams,
		"ingressClass": ingressClass,
	}

	chart, err = kuberlogicIngressControllerChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading kuberlogic ingress controller chart")
	}
	i.Log.Infof("Deploying Kuberlogic Ingress configuration")
	if err := releaseHelmChart(helmKuberlogicIngressChart, i.ReleaseNamespace, chart, kuberlogicIngressValues, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying Kuberlogic Ingress Controller configuration")
	}
	return nil
}

func deployUI(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"config": map[string]interface{}{
			"apiEndpoint":               "http://" + i.Endpoints.API,
			"monitoringConsoleEndpoint": "http://" + i.Endpoints.MonitoringConsole + "/login",
		},
		"ingress": map[string]interface{}{
			"enabled": true,
			"host":    i.Endpoints.UI,
			"class":   ingressClass,
		},
	}

	chart, err := uiChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading ui chart")
	}

	i.Log.Infof("Deploying Kuberlogic UI...")
	release.UpdateUIAddress("http://" + i.Endpoints.UI)
	return releaseHelmChart(helmUIChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployApiserver(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"config": map[string]interface{}{
			"port":      apiserverPort,
			"debugLogs": apiserverDebuglLogsEnabled,
			"auth": map[string]interface{}{
				"provider": apiserverAuthProvider,
				"keycloak": map[string]interface{}{
					"clientId":     keycloakClientId,
					"clientSecret": keycloakClientSecret,
					"realmName":    keycloakRealmName,
					"URL":          keycloaInternalkURL,
				},
			},
		},
		"ingress": map[string]interface{}{
			"enabled": true,
			"host":    i.Endpoints.API,
			"class":   ingressClass,
		},
	}

	chart, err := apiserverChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading apiserver chart")
	}

	i.Log.Infof("Deploying Kuberlogic apiserver...")
	release.UpdateAPIAddress("http://" + i.Endpoints.API)
	return releaseHelmChart(helmApiserverChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployOperator(globals map[string]interface{}, i *HelmInstaller) error {
	values := map[string]interface{}{
		"image": map[string]interface{}{
			"repository": operatorRepository,
		},
		"config": map[string]interface{}{
			"grafana": map[string]interface{}{
				"enabled":                   true,
				"endpoint":                  fmt.Sprintf("http://%s:%d/", grafanaServiceName, grafanaServicePort),
				"secret":                    grafanaSecretName,
				"defaultDatasourceEndpoint": "http://" + victoriaMetricsServiceName,
			},
		},
	}

	chart, err := operatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading operator chart")
	}

	i.Log.Infof("Deploying Kuberlogic operator...")
	return releaseHelmChart(helmOperatorChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployMonitoring(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"victoriametrics": map[string]interface{}{
			"service": map[string]interface{}{
				"name": victoriaMetricsServiceName,
			},
		},
		"grafana": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": grafanaImageRepo,
				"tag":        grafanaImageTag,
			},
			"service": map[string]interface{}{
				"name": grafanaServiceName,
			},
			"secretName": grafanaSecretName,
			"admin": map[string]interface{}{
				"user":     grafanaAdminUser,
				"password": release.InternalPassword(),
			},
			"mysql": map[string]interface{}{
				"enabled":      true,
				"rootPassword": release.InternalPassword(),
			},
			"port": grafanaServicePort,
			"auth": map[string]interface{}{
				"headerName": grafanaAuthHeaderName,
			},
			"ingress": map[string]interface{}{
				"enabled": true,
				"host":    i.Endpoints.MonitoringConsole,
				"class":   ingressClass,
				"grafanaLogin": map[string]interface{}{
					"annotations": map[string]interface{}{
						"konghq.com/plugins": fmt.Sprintf("%s,%s,%s", kongJWT2HeadersPlugin, kongJWTCleanupPlugin, kongJWTAuthPlugin),
					},
				},
			},
		},
	}

	chart, err := monitoringChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading monitoring chart")
	}

	i.Log.Infof("Deploying Kuberlogic monitoring...")
	return releaseHelmChart(helmMonitoringChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployServiceOperators(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	// postgres first
	pgValues := map[string]interface{}{
		"crd": map[string]interface{}{
			"create": true,
		},

		"image": map[string]interface{}{
			"registry":   registryName,
			"repository": postgresImageRepo,
			"tag":        postgresImageTag,
		},

		"configKubernetes": map[string]interface{}{
			"secret_name_template": postgresSecretTemplate,
		},
	}

	pgChart, err := postgresOperatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading postgres chart")
	}

	i.Log.Infof("Deploying postgres operator...")
	if err := releaseHelmChart(postgresOperatorChart, i.ReleaseNamespace, pgChart, pgValues, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying postgres chart")
	}

	mysqlValues := map[string]interface{}{
		"installCRDs": false,

		"image": mysqlImage,

		"orchestrator": map[string]interface{}{
			"ingress": map[string]interface{}{
				"enabled": false,
			},
			"topologyPassword": release.InternalPassword(),
		},

		"podDisruptionBudget": map[string]interface{}{
			"enabled": false,
		},

		"podSecurityPolicy": map[string]interface{}{
			"enabled": false,
		},
	}

	mysqlChart, err := mysqlOperatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading mysql chart")
	}

	i.Log.Infof("Deploying MySQL operator...")
	if err := releaseHelmChart(mysqlOperatorChart, i.ReleaseNamespace, mysqlChart, mysqlValues, globals, i.HelmActionConfig, i.Log); err != nil {
		return err
	}
	return err
}
