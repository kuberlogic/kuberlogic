package k8s

import (
	"flag"
	cloudlinuxv1 "gitlab.com/cloudmanaged/operator/api/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func GetConfig() (*rest.Config, error) {
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String(
			"kubeconfig",
			filepath.Join(home, ".kube", "config"),
			"(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String(
			"kubeconfig",
			"",
			"absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	return config, err
}

func GetBaseClient(config *rest.Config) (*kubernetes.Clientset, error) {

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset, nil
}

func GetCloudmanagedClient(config *rest.Config) (*rest.RESTClient, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &cloudlinuxv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	restClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}
	return restClient, nil

}

//func GetServices(clientset *kubernetes.Clientset, clustername string) ([]v1.Service, error) {
//	services, err := clientset.CoreV1().Services("").
//		List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		return nil, err
//	}
//	var clusterServices []v1.Service
//	for _, svc := range services.Items {
//		value, ok := svc.Labels["mysql.presslabs.org/cluster"]
//		if ok && value == clustername {
//			clusterServices = append(clusterServices, svc)
//		}
//	}
//
//	return clusterServices, nil
//}
//
//func GetSecrets(clientset *kubernetes.Clientset, name string) ([]v1.Secret, error) {
//	secrets, err := clientset.CoreV1().Secrets("").
//		List(context.TODO(), metav1.ListOptions{})
//	if err != nil {
//		return nil, err
//	}
//	var clusterSecrets []v1.Secret
//	for _, secret := range secrets.Items {
//		//value, ok := secret.Labels["cluster-name"]
//		if secret.Name == name {
//			clusterSecrets = append(clusterSecrets, secret)
//		}
//	}
//
//	return clusterSecrets, nil
//}
//
//func GetCredentials(secret v1.Secret) (string, string) {
//	return string(secret.Data["username"]), string(secret.Data["password"])
//}
//
//func GetNodePort(svc v1.Service) (int32, error) {
//	var pgPort int32 = 3306
//	for _, port := range svc.Spec.Ports {
//		if port.Port == pgPort {
//			return port.NodePort, nil
//		}
//	}
//	return 0, fmt.Errorf("%d port not found for service %s", pgPort, svc.Name)
//}
