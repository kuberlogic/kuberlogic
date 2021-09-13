package helm_installer

import (
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"helm.sh/helm/v3/pkg/action"
)

func deployCRDs(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := make(map[string]interface{}, 0)

	chart, err := crdsChartReader()
	if err != nil {
		return fmt.Errorf("error loading CRDs chart: %v", err)
	}

	return releaseHelmChart(helmCRDsChart, ns, chart, values, globals, actConfig, log)
}

func deployAuth(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	keycloakLocalValues := map[string]interface{}{
		"installCRDs": false,
	}

	operatorChart, err := keycloakOperatorChartReader()
	if err != nil {
		return fmt.Errorf("error loading keycloak-operator chart: %vv", err)
	}

	if err := releaseHelmChart(helmKeycloakOperatorChart, ns, operatorChart, keycloakLocalValues, globals, actConfig, log); err != nil {
		return fmt.Errorf("error installing keycloak-operator: %vv", err)
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
		return fmt.Errorf("error loading kuberlogic-keycloak chart: %v", err)
	}
	if err := releaseHelmChart(helmKuberlogicKeycloakCHart, ns, kuberlogicKeycloakChart, kuberlogicKeycloakValues, globals, actConfig, log); err != nil {
		return fmt.Errorf("error installing kuberlogic-keycloak: %v", err)
	}
	return nil
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
		return fmt.Errorf("error loading apiserver chart: %v", err)
	}

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
		return fmt.Errorf("error loading operator chart: %v", err)
	}

	return releaseHelmChart(helmOperatorChart, ns, chart, values, globals, actConfig, log)
}

func deployMonitoring(ns string, globals map[string]interface{}, actConfig *action.Configuration, log logger.Logger) error {
	values := map[string]interface{}{}

	chart, err := monitoringChartReader()
	if err != nil {
		return fmt.Errorf("error loading monitoring chart: %v", err)
	}

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
		return fmt.Errorf("error loading postgres chart: %v", err)
	}

	if err := releaseHelmChart(postgresOperatorChart, ns, pgChart, pgValues, globals, actConfig, log); err != nil {
		return fmt.Errorf("error installing postgres chart: %v", err)
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
		return fmt.Errorf("error loading mysql chart: %v", err)
	}

	if err := releaseHelmChart(mysqlOperatorChart, ns, mysqlChart, mysqlValues, globals, actConfig, log); err != nil {
		return err
	}
	return err
}
