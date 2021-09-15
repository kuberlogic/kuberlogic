package helm_installer

import (
	"context"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

func deployCRDs(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := make(map[string]interface{}, 0)

	chart, err := crdsChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading CRDs chart")
	}

	log.Infof("Deploying CRDs...")
	return releaseHelmChart(helmCRDsChart, ns, chart, values, globals, actConfig, log)
}

func deployCertManager(globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := map[string]interface{}{
		"installCRDs": true,
	}

	chart, err := certManagerChartReader()
	if err != nil {
		errors.Wrap(err, "error loading cert-manager chart")
	}

	log.Infof("Deploying cert-manager...")
	return releaseHelmChart(helmCertManagerChart, certManagerNs, chart, values, globals, actConfig, log)
}

func deployAuth(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger, clientset *kubernetes.Clientset) error {
	keycloakLocalValues := map[string]interface{}{
		"installCRDs": false,
	}

	operatorChart, err := keycloakOperatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading keycloak-operator chart")
	}

	log.Infof("Deploying keycloak-operator...")
	if err := releaseHelmChart(helmKeycloakOperatorChart, ns, operatorChart, keycloakLocalValues, globals, actConfig, log); err != nil {
		return errors.Wrap(err, "error deploying keycloak-operator")
	}

	kuberlogicKeycloakValues := map[string]interface{}{
		"realm": map[string]interface{}{
			"id":   keycloakRealmName,
			"name": keycloakRealmName,
		},
		"clientId":     keycloakClientId,
		"clientSecret": keycloakClientSecret,

		"apiserverId": oauthApiserverId,
	}

	kuberlogicKeycloakChart, err := kuberlogicKeycloakChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading kuberlogic-keycloak chart")
	}
	log.Infof("Deploying Kuberlogic Keycloak resources...")
	if err := releaseHelmChart(helmKuberlogicKeycloakCHart, ns, kuberlogicKeycloakChart, kuberlogicKeycloakValues, globals, actConfig, log); err != nil {
		return errors.Wrap(err, "error deploying kuberlogic-keycloak")
	}
	if err := waitForKeycloakResources(ns, clientset); err != nil {
		return errors.Wrap(err, "keycloak provisioning error")
	}
	return nil
}

func deployNginxIC(ns string, globals map[string]interface{}, actConfig *action.Configuration, clientset *kubernetes.Clientset, log logger.Logger) error {
	values := map[string]interface{}{
		"defaultBackend": map[string]interface{}{
			"enabled": false,
		},
	}
	chart, err := nginxIngressControllerChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading nginx-ingress-controller chart")
	}
	log.Infof("Deploying Nginx Ingress Controller...")
	if err := releaseHelmChart(helmNginxIngressChart, ns, chart, values, globals, actConfig, log); err != nil {
		return errors.Wrap(err, "error deploying nginx-ingress-controller")
	}

	// verify that Nginx Ingress Controller services is created and received Ingress IP address
	// service name equals to the chart name
	const waitTimeoutSec = 30
	for i := 0; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		s, err := clientset.CoreV1().Services(ns).Get(context.TODO(), helmNginxIngressChart, v1.GetOptions{})
		if err != nil {
			continue // hope that the error is transient
		}
		if len(s.Status.LoadBalancer.Ingress) != 0 {
			// success
			return nil
		}
	}
	return errors.New("failed to obtain an Ingress IP address for nginx-ingress-controller")
}

func deployApiserver(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
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
	}

	chart, err := apiserverChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading apiserver chart")
	}

	log.Infof("Deploying Kuberlogic apiserver...")
	return releaseHelmChart(helmApiserverChart, ns, chart, values, globals, actConfig, log)
}

func deployOperator(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := map[string]interface{}{
		"image": map[string]interface{}{
			"tag":        operatorTag,
			"repository": operatorRepository,
		},
		"installCRDs": false,
	}

	chart, err := operatorChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading operator chart")
	}

	log.Infof("Deploying Kuberlogic operator...")
	return releaseHelmChart(helmOperatorChart, ns, chart, values, globals, actConfig, log)
}

func deployMonitoring(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := map[string]interface{}{}

	chart, err := monitoringChartReader()
	if err != nil {
		return errors.Wrap(err, "error loading monitoring chart")
	}

	log.Infof("Deploying Kuberlogic monitoring...")
	return releaseHelmChart(helmMonitoringChart, ns, chart, values, globals, actConfig, log)
}

func deployServiceOperators(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
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

	log.Infof("Deploying postgres operator...")
	if err := releaseHelmChart(postgresOperatorChart, ns, pgChart, pgValues, globals, actConfig, log); err != nil {
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

	log.Infof("Deploying MySQL operator...")
	if err := releaseHelmChart(mysqlOperatorChart, ns, mysqlChart, mysqlValues, globals, actConfig, log); err != nil {
		return err
	}
	return err
}
