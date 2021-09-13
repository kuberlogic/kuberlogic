package helm_installer

import (
	"context"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// secrets are created by Keycloak operator during realm/client provisioning
const (
	clientSecretName = "keycloak-client-secret-apiserver-client"
	realmSecretName  = "credential-kuberlogic-realm-kuberlogic-realm-kuberlogic"
)

func uninstallKuberlogicKeycloak(ns string, force bool, act *action.Configuration, clientset *kubernetes.Clientset, log logger.Logger) error {
	if err := uninstallHelmChart(helmKuberlogicKeycloakCHart, force, act, log); err != nil {
		return err
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted. Waiting for confirmation.")
	for _, c := range []string{clientSecretName, realmSecretName} {
		if err := waitForSecretDeletion(c, ns, clientset); err != nil {
			return fmt.Errorf("failed to clean up Keycloak resources: %s", err)
		}
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted")

	return uninstallHelmChart(helmKeycloakOperatorChart, force, act, log)
}

func waitForSecretDeletion(name, ns string, clientset *kubernetes.Clientset) error {
	const deleteTimeout = 300

	for i := 1; i < deleteTimeout; i *= 2 {
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if errors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		continue
	}
	return fmt.Errorf("secret %s is still present", name)
}