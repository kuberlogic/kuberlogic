package helm_installer

import (
	"context"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"
)

// secrets are created by Keycloak operator during realm/client provisioning
const (
	clientSecretName = "keycloak-client-secret-apiserver-client"
	realmSecretName  = "credential-kuberlogic-realm-kuberlogic-realm-kuberlogic"
)

func waitForKeycloakResources(ns string, clientset *kubernetes.Clientset) error {
	for _, c := range []string{realmSecretName, clientSecretName} {
		if err := waitForSecretCreation(c, ns, clientset); err != nil {
			return errors.Wrap(err, "failed to wait for Keycloak resources")
		}
	}
	return nil
}

func uninstallKuberlogicKeycloak(ns string, force bool, act *action.Configuration, clientset *kubernetes.Clientset, log logger.Logger) error {
	if err := uninstallHelmChart(helmKuberlogicKeycloakCHart, force, act, log); err != nil {
		return err
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted. Waiting for confirmation.")
	for _, c := range []string{clientSecretName, realmSecretName} {
		if err := waitForSecretDeletion(c, ns, clientset); err != nil {
			return errors.Wrap(err, "failed to clean up Keycloak resources")
		}
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted")

	return uninstallHelmChart(helmKeycloakOperatorChart, force, act, log)
}

func waitForSecretCreation(name, ns string, clientset *kubernetes.Clientset) error {
	const minutesWait = 6

	for i := 1; i < minutesWait*2; i += 1 {
		time.Sleep(30 * time.Second)
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			continue
		}
		return nil
	}
	return errors.New("secret is not present :" + name)
}

func waitForSecretDeletion(name, ns string, clientset *kubernetes.Clientset) error {
	const minutesWait = 6

	for i := 1; i < minutesWait*2; i *= 2 {
		time.Sleep(30 * time.Second)
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		continue
	}
	return errors.New("secret is still present: " + name)
}
