package helm_installer

import (
	"context"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func checkKubernetesVersion(clientset *kubernetes.Clientset, log logger.Logger) error {
	log.Infof("Checking Kubernetes cluster version")
	info, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("error checking Kubernetes version: %s", err)
	}
	compatMajor, compatMinor := "1", "20" // this is the only tested Kubernetes version
	if info.Major != compatMajor && info.Minor != compatMinor {
		log.Errorf("Kubernetes version %s.%s is required, but %s found", compatMajor, compatMinor, info.String())
		return fmt.Errorf("cluster version incompatible")
	}
	log.Infof("Kubernetes version %s is found.", info.String())
	return nil
}

func checkDefaultStorageClass(clientset *kubernetes.Clientset, log logger.Logger) error {
	log.Infof("Checking default storage class")
	scs, err := clientset.StorageV1().StorageClasses().List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("error listing storage classes: %v", err)
	}
	found := false
	for _, sc := range scs.Items {
		if anno := sc.GetAnnotations(); anno["storageclass.kubernetes.io/is-default-class"] != "" {
			found = true
			log.Infof("Found default storage class: %s", sc.Name)
		}
	}
	if !found {
		log.Errorf("Default storage class is not found")
		return fmt.Errorf("default storage class is not found")
	}
	return nil
}
