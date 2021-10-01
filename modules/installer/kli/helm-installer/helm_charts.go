package helm_installer

import (
	"embed"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/release"
	"io"
	"time"
)

const (
	helmCRDsChart = "crds"

	helmCertManagerChart = "cert-manager"

	helmKongIngressControllerChart = "kong"
	helmKuberlogicIngressChart     = "kuberlogic-ingress-controller"

	helmKeycloakOperatorChart   = "keycloak-operator"
	helmKuberlogicKeycloakCHart = "kuberlogic-keycloak"
	helmMonitoringChart         = "monitoring"

	postgresOperatorChart = "postgres-operator"
	mysqlOperatorChart    = "mysql-operator"

	helmOperatorChart  = "kuberlogic-operator"
	helmApiserverChart = "kuberlogic-apiserver"
	helmUIChart        = "kuberlogic-ui"

	helmActionTimeoutSec = 300
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
	//go:embed kuberlogic-ui-0.1.0.tgz
	//go:embed cert-manager-v1.3.1.tgz
	//go:embed kong-2.3.0.tgz
	//go:embed kuberlogic-ingress-controller-0.1.0.tgz
	helmFs embed.FS
)

func kuberlogicIngressControllerChartReader() (io.Reader, error) {
	return helmFs.Open("kuberlogic-ingress-controller-0.1.0.tgz")
}

func kongIngressControllerChartReader() (io.Reader, error) {
	return helmFs.Open("kong-2.3.0.tgz")
}

func certManagerChartReader() (io.Reader, error) {
	return helmFs.Open("cert-manager-v1.3.1.tgz")
}

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
	return helmFs.Open("kuberlogic-operator-0.1.0.tgz")
}

func apiserverChartReader() (io.Reader, error) {
	return helmFs.Open("kuberlogic-apiserver-0.1.0.tgz")
}

func uiChartReader() (io.Reader, error) {
	return helmFs.Open("kuberlogic-ui-0.1.0.tgz")
}

func findHelmRelease(name string, c *action.Configuration, log logger.Logger) (*release.Release, error) {
	list := action.NewList(c)
	list.All = true
	list.StateMask =
		action.ListDeployed |
			action.ListUninstalling |
			action.ListPendingInstall |
			action.ListPendingUpgrade |
			action.ListPendingRollback |
			action.ListFailed |
			action.ListAll

	releases, err := list.Run()
	if err != nil {
		return nil, err
	}

	log.Debugf("List of installed Helm releases: %v", releases)
	for _, r := range releases {
		if r.Name == name {
			return r, nil
		}
	}
	return nil, nil
}

// releaseHelmChart computes values for Helm Chart and upgrades it if it is already installed
// otherwise it installs it
func releaseHelmChart(name, ns string, chartReader io.Reader, locals, globals map[string]interface{}, conf *action.Configuration, log logger.Logger) error {
	// load chart
	c, err := loader.LoadArchive(chartReader)
	if err != nil {
		return fmt.Errorf("error loading chart archive: %v", err)
	}

	resultVals, err := mergeValues(globals, locals)
	if err != nil {
		return fmt.Errorf("error computing values for a chart: %v", err)
	}
	log.Debugf("Releasing %s with values", name, resultVals)

	// search for already installed release
	r, err := findHelmRelease(name, conf, log)
	if err != nil {
		return errors.Wrap(err, "error releasing chart")
	}

	// release exists and it is in pending state
	if r != nil && r.Info.Status.IsPending() {
		uninstallHelmChart(name, true, conf, log)
		r = nil
	}

	// Helm Chart release error
	var releaseErr error
	// Signature of the installation func. It is the same for upgrade/install processes
	var releaseFunc func(string, string, *chart.Chart, map[string]interface{}, *action.Configuration, logger.Logger) error
	if r == nil {
		releaseFunc = installHelmChart
	} else {
		releaseFunc = upgradeHelmChart
	}

	// retry in case of errors
	maxRetries := 3
	for i := 0; i < maxRetries; i += 1 {
		if releaseErr = releaseFunc(name, ns, c, resultVals, conf, log); releaseErr == nil {
			return nil
		}
		log.Debugf("Error happened: %s. Retries left: %d", releaseErr.Error(), i)
		time.Sleep(time.Second * time.Duration(i))
	}
	return releaseErr
}

// upgradeHelmChart upgrades a Helm chart with given values
func upgradeHelmChart(name, ns string, chart *chart.Chart, values map[string]interface{}, c *action.Configuration, log logger.Logger) error {
	// create install action
	action := action.NewUpgrade(c)
	action.Install = true
	action.Namespace = ns
	action.SkipCRDs = false
	action.Timeout = time.Second * helmActionTimeoutSec
	action.Wait = true
	log.Debugf("Upgrade action configuration: %+v", action)

	rel, err := action.Run(name, chart, values)
	log.Debugf("Upgrade error for chart %s release %+v: %+v", name, rel, err)
	if err != nil {
		return fmt.Errorf("error upgrading %s:  %v", name, err)
	}
	log.Infof("%s successfully upgraded. Status: %+v\n", name, rel.Info.Status)
	return nil
}

// installHelmChart installs a Helm chart with name `name` into a namespace `ns`
func installHelmChart(name, ns string, chart *chart.Chart, values map[string]interface{}, c *action.Configuration, log logger.Logger) error {
	// create install action
	installAction := action.NewInstall(c)
	installAction.Atomic = true
	installAction.Wait = true
	installAction.Timeout = time.Second * helmActionTimeoutSec
	installAction.Namespace = ns
	installAction.CreateNamespace = true
	installAction.ReleaseName = name
	installAction.IncludeCRDs = true
	installAction.SkipCRDs = false
	log.Debugf("Install action configuration: %+v", installAction)

	rel, installErr := installAction.Run(chart, values)
	log.Debugf("installation error for chart %s release %+v: %+v", name, rel, installErr)
	if installErr != nil {
		return fmt.Errorf("error installing %s:  %v", name, installErr)
	}
	log.Infof("%s successfully installed. Status: %+v\n", name, rel.Info.Status)
	return nil
}

// uninstallHelmChart uninstalls a Helm Release with name `name`
func uninstallHelmChart(name string, force bool, actConfig *action.Configuration, log logger.Logger) error {
	release, err := findHelmRelease(name, actConfig, log)
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
