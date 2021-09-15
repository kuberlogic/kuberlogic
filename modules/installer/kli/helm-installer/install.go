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

	// prepare environment for release and start release process
	if err := internal.PrepareEnvironment(i.ReleaseNamespace, i.Registry.Server, i.Registry.Password, i.Registry.Username, i.ClientSet); err != nil {
		return errors.Wrap(err, "error preparing environment")
	}
	if _, err := internal.StartRelease(i.ReleaseNamespace, i.ClientSet); err != nil {
		return errors.Wrap(err, "error starting release")
	}

	err := func() error {
		// install CRDs into cluster
		i.Log.Infof("Installing CRDs...")
		if err := deployCRDs(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
			return errors.Wrap(err, "error installing CRDs")
		}

		if installPhase == "all" || installPhase == "dependencies" {
			i.Log.Infof("Installing Kuberlogic dependencies...")
			if err := deployNginxIC(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.ClientSet, i.Log); err != nil {
				return errors.Wrap(err, "error installing nginx-ingress-controller")
			}
			if err := deployCertManager(globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing cert-manager")
			}

			if err := deployAuth(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log, i.ClientSet); err != nil {
				return errors.Wrap(err, "error installing keycloak")
			}

			if err := deployServiceOperators(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing service operators")
			}

			if err := deployMonitoring(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing monitoring component")
			}
		}

		if installPhase == "all" || installPhase == "kuberlogic" {
			i.Log.Infof("Installing Kuberlogic core components...")
			if err := deployOperator(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing operator")
			}

			if err := deployApiserver(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing apiserver")
			}

			if err := deployUI(i.ReleaseNamespace, globalValues, i.HelmActionConfig, i.Log); err != nil {
				return errors.Wrap(err, "error installing UI")
			}
		}
		return nil
	}()
	if err != nil {
		i.Log.Infof("Installation failed: %v", err)
		internal.FailRelease(i.ReleaseNamespace, i.ClientSet)
		return err
	}
	_, err = internal.FinishRelease(i.ReleaseNamespace, i.ClientSet)
	i.Log.Infof("Installation completed successfully!")
	return err
}

func runInstallChecks(clientSet *kubernetes.Clientset, actionConfig *action.Configuration, log logger.Logger) error {
	if err := checkKubernetesVersion(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes version")
	}
	if err := checkDefaultStorageClass(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes default StorageClass")
	}
	if err := checkLoadBalancerServiceType(clientSet, log); err != nil {
		return errors.Wrap(err, "error checking Kubernetes LoadBalancer service")
	}
	return nil
}
