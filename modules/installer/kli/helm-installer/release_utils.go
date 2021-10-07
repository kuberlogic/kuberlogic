package helm_installer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	logger "github.com/kuberlogic/kuberlogic/modules/installer/log"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"io"
	v12 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"time"
)

// secrets are created by Keycloak operator during realm/client provisioning
const (
	realmSecretName = "credential-kuberlogic-realm-kuberlogic-realm-kuberlogic"
)

func getJWTAuthVals(ns string, clientset *kubernetes.Clientset, log logger.Logger) (map[string]interface{}, error) {
	// get JWKS info from keycloak and save as a secret
	log.Debugf("getting keycloak nodeport service")
	svc, err := clientset.CoreV1().Services(ns).Get(context.TODO(), keycloakNodePortServiceName, v1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error getting Keycloak service")
	}
	nodePort := svc.Spec.Ports[0].NodePort
	log.Debugf("keycloak nodeport service port is %d", nodePort)
	if nodePort == 0 {
		return nil, errors.New("nodePort for Keycloak service can't be empty")
	}
	// iterate over Kubernetes nodes until we can reach Keycloak via nodePort
	log.Debugf("getting Kubernetes nodes")
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "error listing Kubernetes nodes")
	}
	if len(nodes.Items) == 0 {
		return nil, errors.New("Kubernetes can't have zero nodes")
	}

	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 5,
	}

	// iterate over all node/address combinations in cluster until we reach keycloak via nodePort
	// usuallu it will succeed on 1st or 2nd address attempt because nodePort is opened on all cluster nodes
	log.Debugf("getting realm public key data from Keycloak via nodePort service")
	keycloakRealmAuthData := make([]byte, 0)
found:
	for _, n := range nodes.Items {
		for _, addr := range n.Status.Addresses {
			if addr.Type == v12.NodeHostName {
				continue // skip hostname
			}
			endpoint := fmt.Sprintf("https://%s:%d/auth/realms/%s",
				addr.Address, nodePort, keycloakRealmName)

			log.Debugf("trying %s endpoint via node %s", endpoint, n.Name)
			r, err := httpClient.Get(endpoint)
			if err != nil {
				continue // try next address/node combination
			}
			keycloakRealmAuthData, err = io.ReadAll(r.Body)
			if err != nil {
				continue // try next address/node combination
			}
			r.Body.Close()
			log.Debugf("keycloak realm auth data request response is: %s", string(keycloakRealmAuthData))
			if r.StatusCode != http.StatusOK {
				return nil, errors.New("keycloak answered with non ok status code")
			}
			break found
		}
	}
	// no valid response received
	if len(keycloakRealmAuthData) == 0 {
		return nil, errors.New("failed getting realm auth data from Keycloak")
	}
	var authData struct {
		PublicKey string `json:"public_key"`
	}
	if err := json.Unmarshal(keycloakRealmAuthData, &authData); err != nil {
		return nil, errors.Wrap(err, "error decoding Keycloak response")
	}

	return map[string]interface{}{
		"name":      ingressAuthName,
		"param":     jwtTokenQueryParam,
		"iss":       jwtIssuer,
		"rsaPubKey": "-----BEGIN PUBLIC KEY-----\n" + authData.PublicKey + "\n-----END PUBLIC KEY-----\n",
	}, nil
}

func waitForKeycloakResources(ns string, clientset *kubernetes.Clientset) error {
	if err := waitForSecretCreation(realmSecretName, ns, clientset); err != nil {
		return errors.Wrap(err, "failed to wait for Keycloak resources")
	}
	return nil
}

func uninstallKuberlogicKeycloak(ns string, force bool, act *action.Configuration, clientset *kubernetes.Clientset, log logger.Logger) error {
	if err := uninstallHelmChart(helmKuberlogicKeycloakCHart, force, act, log); err != nil {
		return err
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted. Waiting for confirmation.")
	if err := waitForSecretDeletion(realmSecretName, ns, clientset); err != nil {
		return errors.Wrap(err, "failed to clean up Keycloak resources")
	}
	log.Debugf("Kuberlogic Keycloak resources are deleted")

	return uninstallHelmChart(helmKeycloakOperatorChart, force, act, log)
}

func waitForSecretCreation(name, ns string, clientset *kubernetes.Clientset) error {
	const waitTimeoutSec = 450

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		_, err := clientset.CoreV1().Secrets(ns).Get(context.TODO(), name, v1.GetOptions{})
		if err != nil {
			continue
		}
		return nil
	}
	return errors.New("secret is not present :" + name)
}

func waitForSecretDeletion(name, ns string, clientset *kubernetes.Clientset) error {
	const waitTimeoutSec = 450

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
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
