package helm_installer

import (
	"context"
	"github.com/kuberlogic/operator/modules/installer/internal"
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
		errors.Wrap(err, "error loading cert-manager chart")
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
			"id":   keycloakRealmName,
			"name": keycloakRealmName,
		},
		"clientId":     keycloakClientId,
		"clientSecret": keycloakClientSecret,

		"apiserverId":   oauthApiserverId,
		"adminPassword": i.Auth.AdminPassword,
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

func deployNginxIC(globals map[string]interface{}, i *HelmInstaller, releaseInfo *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"defaultBackend": map[string]interface{}{
			"enabled": false,
		},
		"ingressClass": ingressClass,
	}
	chart, err := nginxIngressControllerChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading nginx-ingress-controller chart")
	}
	i.Log.Infof("Deploying Nginx Ingress Controller...")
	if err := releaseHelmChart(helmNginxIngressChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error deploying nginx-ingress-controller")
	}

	// verify that Nginx Ingress Controller services is created and received Ingress IP address
	// service name equals to the chart name
	const waitTimeoutSec = 30
	for x := 0; x < waitTimeoutSec; x += 1 {
		time.Sleep(time.Second)
		s, err := i.ClientSet.CoreV1().Services(i.ReleaseNamespace).Get(context.TODO(), helmNginxIngressChart, v1.GetOptions{})
		if err != nil {
			continue // hope that the error is transient
		}
		if len(s.Status.LoadBalancer.Ingress) != 0 {
			// success. append to the release banner
			releaseInfo.AddBannerLines("Connection endpoint address: " + s.Status.LoadBalancer.Ingress[0].IP)
			return nil
		}
	}
	return errors.New("failed to obtain an Ingress IP address for nginx-ingress-controller")
}

func deployUI(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"config": map[string]interface{}{
			"apiEndpoint": "http://" + i.Endpoints.API,
		},
		"image": map[string]interface{}{
			"tag": uiImageTag,
		},
		"ingress": map[string]interface{}{
			"enabled": true,
			"host":    i.Endpoints.UI,
			"class":   ingressClass,
		},
	}

	chart, err := uiChartReader()
	if err != nil {
		errors.Wrap(err, "error loading ui chart")
	}

	i.Log.Infof("Deploying Kuberlogic UI...")
	release.AddBannerLines("Kuberlogic Web UI endpoint: http://" + i.Endpoints.UI)
	return releaseHelmChart(helmUIChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployApiserver(globals map[string]interface{}, i *HelmInstaller, release *internal.ReleaseInfo) error {
	values := map[string]interface{}{
		"image": map[string]interface{}{
			"tag": apiserverTag,
		},
		"config": map[string]interface{}{
			"port":      apiserverPort,
			"debugLogs": apiserverDebuglLogsEnabled,
			"auth": map[string]interface{}{
				"provider": apiserverAuthProvider,
				"keycloak": map[string]interface{}{
					"clientId":     keycloakClientId,
					"clientSecret": keycloakClientSecret,
					"realmName":    keycloakRealmName,
					"URL":          keycloakURL,
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
	release.AddBannerLines("Kuberlogic API endpoint: http://" + i.Endpoints.API)
	return releaseHelmChart(helmApiserverChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployOperator(globals map[string]interface{}, i *HelmInstaller) error {
	values := map[string]interface{}{
		"image": map[string]interface{}{
			"tag":        operatorTag,
			"repository": operatorRepository,
		},
		"grafana": map[string]interface{}{
			"enabled": false,
		},
	}

	chart, err := operatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading operator chart")
	}

	i.Log.Infof("Deploying Kuberlogic operator...")
	return releaseHelmChart(helmOperatorChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployMonitoring(globals map[string]interface{}, i *HelmInstaller) error {
	values := map[string]interface{}{}

	chart, err := monitoringChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading monitoring chart")
	}

	i.Log.Infof("Deploying Kuberlogic monitoring...")
	return releaseHelmChart(helmMonitoringChart, i.ReleaseNamespace, chart, values, globals, i.HelmActionConfig, i.Log)
}

func deployServiceOperators(globals map[string]interface{}, i *HelmInstaller) error {
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
