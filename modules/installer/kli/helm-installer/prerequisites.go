package helm_installer

import (
	"context"
	"encoding/base64"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	pullSecretName = "kuberlogic-registry"
)

func createPullSecret(name, ns, server, username, password string, clientset *kubernetes.Clientset, log logger.Logger) error {
	sec := &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
		Type: "kubernetes.io/dockerconfigjson",
		Data: map[string][]byte{
			".dockerconfigjson": []byte(
				fmt.Sprintf("{\"auths\":{\"%s\":{\"username\":\"%s\",\"password\":\"%s\",\"auth\":\"%s\"}}}",
					server, username, password, base64.StdEncoding.EncodeToString([]byte(username+":"+password)))),
		},
	}
	log.Debugf("image pull secret to be created: %+v", sec)

	// list secrets
	sl, err := clientset.CoreV1().Secrets(ns).List(context.TODO(), v12.ListOptions{})
	if err != nil {
		return err
	}
	for _, s := range sl.Items {
		if s.Name == name && s.Namespace == ns {
			return fmt.Errorf("secret with the same name already exists: %s/%s", ns, name)
		}
	}
	createdSec, err := clientset.CoreV1().Secrets(ns).Create(context.TODO(), sec, v12.CreateOptions{})
	log.Debugf("image pull secret is created: %+v", createdSec)
	return err
}

func deletePullSecret(name, ns string, force bool, clientset *kubernetes.Clientset) error {
	err := clientset.CoreV1().Secrets(ns).Delete(context.TODO(), name, v12.DeleteOptions{})
	if errors.IsNotFound(err) && force {
		return nil
	}
	return err
}
