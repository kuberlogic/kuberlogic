/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	kuberlogicv1 "github.com/kuberlogic/kuberlogic/modules/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

var kuberLogicAlertCR = "kuberlogicalerts"

func newKubernetesClients() (kubernetes.Interface, rest.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	err = kuberlogicv1.AddToScheme(k8scheme.Scheme)
	if err != nil {
		return nil, nil, err
	}

	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &kuberlogicv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	kubeRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		return nil, nil, err
	}
	kubeClientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}
	return kubeClientSet, kubeRestClient, err
}

func createAlertCR(name, namespace, alertName, alertValue, cluster, pod, summary string, kubeRestClient rest.Interface) error {
	klAlert := kuberlogicv1.KuberLogicAlert{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: kuberlogicv1.KuberLogicAlertSpec{
			AlertName:  alertName,
			AlertValue: alertValue,
			Cluster:    cluster,
			Pod:        pod,
			Summary:    summary,
		},
	}

	return kubeRestClient.Post().
		Namespace(namespace).
		Resource(kuberLogicAlertCR).
		Body(&klAlert).
		Do(context.TODO()).Error()
}

func deleteAlertCR(name, namespace string, kubeRestClient rest.Interface) error {
	return kubeRestClient.Delete().
		Name(name).
		Namespace(namespace).
		Resource(kuberLogicAlertCR).
		Do(context.TODO()).Error()
}
