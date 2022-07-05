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

package k8s

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/config"
	cloudlinuxv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	k8scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func GetConfig(cfg *config.Config) (*rest.Config, error) {
	// check in-cluster usage
	if cfg, err := rest.InClusterConfig(); err == nil {
		return cfg, nil
	}

	// use the current context in kubeconfig
	conf, err := clientcmd.BuildConfigFromFlags("",
		cfg.KubeconfigPath)
	if err != nil {
		return nil, err
	}
	return conf, err
}

func GetKuberLogicClient(config *rest.Config) (rest.Interface, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &cloudlinuxv1alpha1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(k8scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	return rest.UnversionedRESTClientFor(&crdConfig)
}
