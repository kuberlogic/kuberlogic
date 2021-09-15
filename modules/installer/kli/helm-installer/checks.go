package helm_installer

import (
	"context"
	"fmt"
	logger "github.com/kuberlogic/operator/modules/installer/log"
	"github.com/pkg/errors"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

var (
	// compatible kubernetes version in major.minor form
	kubeCompatVersions = []string{"1.20"}
)

// checkKubernetesVersion checks if Kubernetes cluster is ex
func checkKubernetesVersion(clientset *kubernetes.Clientset, log logger.Logger) error {
	log.Infof("Checking Kubernetes cluster version")
	info, err := clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("error checking Kubernetes version: %s", err)
	}
	for _, v := range kubeCompatVersions {
		ver := strings.Split(v, ".")
		if info.Major == ver[0] && info.Minor == ver[1] {
			log.Infof("Kubernetes version %s is found", info.String())
			return nil
		}
	}
	log.Errorf("Kubernetes version %s found", info.String())
	return fmt.Errorf("cluster version incompatible")
}

// checkDefaultStorageClass checks if a Kubernetes cluster has a default StorageClass
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

func checkLoadBalancerServiceType(clientset *kubernetes.Clientset, log logger.Logger) error {
	const (
		testServiceName = "kubernetes-test-service"
		testNamespace   = "default"
		testPort        = 9999
		waitTimeoutSec  = 30
	)
	log.Infof("Checking LoadBalancer service type")
	svc := &v12.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      testServiceName,
			Namespace: testNamespace,
		},
		Spec: v12.ServiceSpec{
			Type: "LoadBalancer",
			Ports: []v12.ServicePort{
				{
					Name:     "test",
					Port:     testPort,
					Protocol: "TCP",
				},
			},
		},
	}

	if _, err := clientset.CoreV1().Services(testNamespace).Create(context.TODO(), svc, v1.CreateOptions{}); err != nil {
		return errors.Wrap(err, "error creating test service")
	}
	log.Debugf("test service %s/%s created", testNamespace, testServiceName)

	defer func() {
		clientset.CoreV1().Services(testNamespace).Delete(context.TODO(), testServiceName, v1.DeleteOptions{})
		log.Debugf("test service %s deleted", testNamespace, testServiceName)
	}()

	for i := 1; i < waitTimeoutSec; i += 1 {
		time.Sleep(time.Second)
		s, err := clientset.CoreV1().Services(testNamespace).Get(context.TODO(), testServiceName, v1.GetOptions{})
		if err != nil {
			return errors.Wrap(err, "error getting test service")
		}
		if len(s.Status.LoadBalancer.Ingress) != 0 {
			log.Infof("Kubernetes cluster supports LoadBalancer Service type")
			return nil
		}
		continue
	}
	return errors.New("Service with LoadBalancer didn't get ingress address")
}
