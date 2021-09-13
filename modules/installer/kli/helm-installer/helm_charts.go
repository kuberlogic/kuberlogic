package helm_installer

import (
	"embed"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"time"
)

const (
	helmCRDsChart = "crds"

	helmKeycloakOperatorChart   = "keycloak-operator"
	helmKuberlogicKeycloakCHart = "kuberlogic-keycloak"
	helmMonitoringChart         = "monitoring"

	postgresOperatorChart = "postgres-operator"
	mysqlOperatorChart    = "mysql-operator"

	helmOperatorChart  = "operator"
	helmApiserverChart = "apiserver"
	helmUIChart        = "ui"

	helmActionTimeoutSec = 600
)

var (
	//go:embed crds-0.1.0.tgz
	//go:embed keycloak-operator-0.1.0.tgz
	//go:embed kuberlogic-keycloak-0.1.0.tgz
	//go:embed monitoring-0.1.0.tgz
	//go:embed mysql-operator-0.1.1+master.tgz
	//go:embed postgres-operator-1.6.2.tgz
	//go:embed kuberlogic-operator-0.1.0.tgz
	//go:embed kuberlogic-apiserver-0.1.0.tgz
	// //go:embed ui-0.1.0.tgz
	helmFs embed.FS
)

func crdsChartReader() (io.Reader, error) {
	return helmFs.Open("crds-0.1.0.tgz")
}

func keycloakOperatorChartReader() (io.Reader, error) {
	return helmFs.Open("keycloak-operator-0.1.0.tgz")
}

func kuberlogicKeycloakChartReader() (io.Reader, error) {
	return helmFs.Open("kuberlogic-keycloak-0.1.0.tgz")
}

func monitoringChartReader() (io.Reader, error) {
	return helmFs.Open("monitoring-0.1.0.tgz")
}

func mysqlOperatorChartReader() (io.Reader, error) {
	return helmFs.Open("mysql-operator-0.1.1+master.tgz")
}

func postgresOperatorChartReader() (io.Reader, error) {
	return helmFs.Open("postgres-operator-1.6.2.tgz")
}

func operatorChartReader() (io.Reader, error) {
	return helmFs.Open("operator-0.1.0.tgz")
}

func apiserverChartReader() (io.Reader, error) {
	return helmFs.Open("apiserver-0.1.0.tgz")
}

func findHelmRelease(name string, c *action.Configuration) (*release.Release, error) {
	list := action.NewList(c)
	list.All = true

	releases, err := list.Run()
	if err != nil {
		return nil, err
	}
	for _, r := range releases {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}

func installHelmChart(name, ns string, chartReader io.Reader, locals, globals map[string]interface{}, c *action.Configuration, log logger.Logger) error {
	// load chart
	chart, err := loader.LoadArchive(chartReader)
	if err != nil {
		return fmt.Errorf("error loading chart archive: %v", err)
	}

	resultVals, err := mergeValues(globals, locals)
	if err != nil {
		return fmt.Errorf("error computing values for a chart: %v", err)
	}

	// create install action
	installAction := action.NewUpgrade(c)
	installAction.Wait = true
	installAction.Timeout = time.Second * helmActionTimeoutSec
	installAction.Namespace = ns
	installAction.Install = true
	installAction.SkipCRDs = false
	log.Debugf("Install action configuration: %+v", installAction)

	log.Debugf("Installing %s with values", name, resultVals)
	rel, installErr := installAction.Run(name, chart, resultVals)
	log.Debugf("installation error for chart %s release %+v: %+v", name, rel, installErr)
	if installErr != nil {
		return fmt.Errorf("error installing %s:  %v", name, installErr)
	}
	log.Infof("%s successfully installed. Status: %+v\n", name, rel.Info.Status)
	return nil
}

func uninstallHelmChart(name string, force bool, actConfig *action.Configuration, log logger.Logger) error {
	release, err := findHelmRelease(name, actConfig)
	if err != nil {
		return err
	}
	// release is not found
	if release == nil {
		log.Debugf("Release %s is not found", name)

		if force {
			return nil
		}
		return fmt.Errorf("release is not found")
	}

	deleteAction := action.NewUninstall(actConfig)

	resp, err := deleteAction.Run(name)
	log.Debugf("Helm action response: %+v", resp)
	if err != nil {
		return fmt.Errorf("error uninstalling %s: %v", name, err)
	}
	return nil
}

// mergeValues returns a new map with values from v1 and v2
func mergeValues(v1 map[string]interface{}, v2 map[string]interface{}) (map[string]interface{}, error) {
	v := make(map[string]interface{}, len(v1)+len(v2))

	for key, val := range v1 {
		if _, found := v[key]; found {
			return v, fmt.Errorf("duplicate key %s found", key)
		}
		v[key] = val
	}
	for key, val := range v2 {
		if _, found := v[key]; found {
			return v, fmt.Errorf("duplicate key %s found", key)
		}
		v[key] = val
	}
	return v, nil
}
