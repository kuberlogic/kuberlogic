package main

import (
	"context"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"log"
)

var cloudmanagedAlertCR = "cloudmanagedalerts"
var kubeRestClient *rest.RESTClient

func initKubernetesClient() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("Error initializing Kubernetes client: %s", err)
	}

	err = cloudlinuxv1.AddToScheme(k8scheme.Scheme)
	if err != nil {
		log.Fatalf("Error adding clientset types to schema! %s", err)
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &cloudlinuxv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	kubeRestClient, err = rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		log.Fatalf("Error initializing Kubernetes client: %s", err)
	}
}

func createAlertCR(name, namespace, alertName, alertValue, cluster, pod string) error {
	cloudmanagedAlert := cloudlinuxv1.CloudManagedAlert{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: cloudlinuxv1.CloudManagedAlertSpec{
			AlertName:  alertName,
			AlertValue: alertValue,
			Cluster:    cluster,
			Pod:        pod,
		},
	}

	res := kubeRestClient.Post().
		Namespace(namespace).
		Resource(cloudmanagedAlertCR).
		Body(&cloudmanagedAlert).
		Do(context.TODO())
	return res.Error()
}

func deleteAlertCR(name, namespace string) error {
	res := kubeRestClient.Delete().
		Name(name).
		Namespace(namespace).
		Resource(cloudmanagedAlertCR).
		Do(context.TODO())
	return res.Error()
}
