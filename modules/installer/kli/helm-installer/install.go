package helm_installer

import (
	"github.com/kuberlogic/operator/modules/installer/internal"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/client-go/kubernetes"
)

func (i *HelmInstaller) Install(args []string) error {
	i.Log.Debugf("entering install phase with args: %+v", args)

	// for now we only expect single arg = see cmd/install.go
	installPhase := args[0]

	// run pre install checks
	if err := runInstallChecks(i.ClientSet, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "pre-install checks are failed")
	}

	// create metadata: release information and image pull secret
	if _, err := internal.StartRelease(i.ReleaseNamespace, i.ClientSet); err != nil {
		return errors.Wrap(err, "error starting release")
	}
	if err := internal.PrepareEnvironment(i.ReleaseNamespace, i.Registry.Server, i.Registry.Password, i.Registry.Username, i.ClientSet); err != nil {
		return errors.Wrap(err, "error creating image pull secret")
	}

	// install CRDs into cluster
	i.Log.Infof("Installing CRDs")
	if err := deployCRDs(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
		return errors.Wrap(err, "error installing CRDs")
	}

	if installPhase == "all" || installPhase == "dependencies" {
		i.Log.Infof("Installing authentication component")
		if err := deployAuth(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing keycloak")
		}

		i.Log.Infof("Installing service operators")
		if err := deployServiceOperators(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing service operators")
		}

		i.Log.Infof("Installing monitoring component")
		if err := deployMonitoring(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing monitoring component")
		}
	}

	if installPhase == "all" || installPhase == "kuberlogic" {
		i.Log.Infof("Installing operator")
		if err := deployOperator(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing operator")
		}

		i.Log.Infof("Installing apiserver")
		if err := deployApiserver(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing apiserver")
		}
	}

	_, err := internal.FinishRelease(i.ReleaseNamespace, i.ClientSet)
	i.Log.Infof("Installation completed successfully!")
	return err
}

func runInstallChecks(clientSet *kubernetes.Clientset, actionConfig *action.Configuration, log logger.Logger) error {
	if err := checkKubernetesVersion(clientSet, log); err != nil {
		return err
	}
	if err := checkDefaultStorageClass(clientSet, log); err != nil {
		return err
	}
	return nil
}
