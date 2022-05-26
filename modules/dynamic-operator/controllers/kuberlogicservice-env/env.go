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

package kuberlogicservice_env

import (
	"context"
	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v13 "k8s.io/api/networking/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Environment struct {
	NamespaceName string
}

// SetupEnv checks if KLS environment is present and creates it if it is not
func SetupEnv(kls *kuberlogiccomv1alpha1.KuberLogicService, c client.Client, ctx context.Context) (*Environment, error) {
	ns := getNamespace(kls)
	if _, err := controllerruntime.CreateOrUpdate(ctx, c, ns, func() error {
		return controllerruntime.SetControllerReference(kls, ns, c.Scheme())
	}); err != nil {
		return nil, errors.Wrap(err, "error setting up kls namespace")
	}

	netpol := getNetworkPolicy(kls, ns)
	if _, err := controllerruntime.CreateOrUpdate(ctx, c, netpol, func() error {
		return controllerruntime.SetControllerReference(kls, netpol, c.Scheme())
	}); err != nil {
		return nil, errors.Wrap(err, "error setting up kls networkpolicy")
	}

	// set namespace status field
	kls.Status.Namespace = ns.GetName()
	return &Environment{
		NamespaceName: ns.GetName(),
	}, nil
}

func getNamespace(kls *kuberlogiccomv1alpha1.KuberLogicService) *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: v12.ObjectMeta{
			Name:   kls.GetName(),
			Labels: envLabels(kls),
		},
	}
}

func getNetworkPolicy(kls *kuberlogiccomv1alpha1.KuberLogicService, ns *v1.Namespace) *v13.NetworkPolicy {
	return &v13.NetworkPolicy{
		ObjectMeta: v12.ObjectMeta{
			Name:      "namespace-only-egress",
			Namespace: ns.Name,
			Labels:    envLabels(kls),
		},
		Spec: v13.NetworkPolicySpec{
			PolicyTypes: []v13.PolicyType{
				v13.PolicyTypeEgress,
			},
			Egress: []v13.NetworkPolicyEgressRule{
				{
					To: []v13.NetworkPolicyPeer{
						{
							NamespaceSelector: &v12.LabelSelector{
								MatchLabels: envLabels(kls),
							},
						},
					},
				},
			},
		},
	}
}

func envLabels(kls *kuberlogiccomv1alpha1.KuberLogicService) map[string]string {
	return map[string]string{
		"kuberlogic.com/env": kls.Name,
	}
}
