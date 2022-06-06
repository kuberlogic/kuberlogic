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
	config "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	v13 "k8s.io/api/networking/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

type EnvironmentManager struct {
	client.Client

	NamespaceName string

	kls *kuberlogiccomv1alpha1.KuberLogicService
	cfg *config.Config
	ctx context.Context
}

// SetupEnv checks if KLS environment is present and creates it if it is not
func SetupEnv(kls *kuberlogiccomv1alpha1.KuberLogicService, c client.Client, cfg *config.Config, ctx context.Context) (*EnvironmentManager, error) {
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

	tlsSecret := &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      cfg.SvcOpts.TLSSecretName,
			Namespace: ns.GetName(),
		},
	}
	if cfg.SvcOpts.TLSSecretName != "" {
		if _, err := controllerruntime.CreateOrUpdate(ctx, c, tlsSecret, func() error {
			srcSecret := &v1.Secret{
				ObjectMeta: v12.ObjectMeta{
					Name:      cfg.SvcOpts.TLSSecretName,
					Namespace: cfg.Namespace,
				},
			}
			if err := c.Get(ctx, client.ObjectKeyFromObject(srcSecret), srcSecret); err != nil {
				return errors.Wrap(err, "error getting source TLS secret")
			}
			tlsSecret.Data = srcSecret.Data
			return controllerruntime.SetControllerReference(kls, tlsSecret, c.Scheme())
		}); err != nil {
			return nil, errors.Wrap(err, "error syncing TLS Secret")
		}
	}

	// set namespace status field
	kls.Status.Namespace = ns.GetName()
	return &EnvironmentManager{
		Client: c,

		NamespaceName: ns.GetName(),

		cfg: cfg,
		kls: kls,
		ctx: ctx,
	}, nil
}

// ExposeService takes internal service name and creates an Ingress object exposing it to the outside
func (e *EnvironmentManager) ExposeService(svcName string, provisionIngress bool) (string, error) {
	// get service
	svc := &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:      svcName,
			Namespace: e.NamespaceName,
		},
	}
	if err := e.Get(e.ctx, client.ObjectKeyFromObject(svc), svc); err != nil {
		return "", errors.Wrapf(err, "error getting service %s/%s", e.NamespaceName, svcName)
	}

	if provisionIngress {
		return e.syncIngress(svc)
	}
	return e.syncLoadBalancer(svc)
}

func (e *EnvironmentManager) syncLoadBalancer(svc *v1.Service) (string, error) {
	port := svc.Spec.Ports[0].Port

	var addr string
	if len(svc.Spec.ExternalIPs) > 0 {
		addr = svc.Spec.ExternalIPs[0]
	} else if svc.Spec.LoadBalancerIP != "" {
		addr = svc.Spec.LoadBalancerIP
	} else if lbIngress := svc.Status.LoadBalancer.Ingress; len(lbIngress) > 0 {
		addr = lbIngress[0].IP
		port = lbIngress[0].Ports[0].Port
	} else {
		return "", nil
	}
	return addr + ":" + strconv.Itoa(int(port)), nil
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
func (e *EnvironmentManager) syncIngress(svc *v1.Service) (string, error) {
	ingress := &v13.Ingress{
		ObjectMeta: v12.ObjectMeta{
			Name:      e.kls.GetName(),
			Namespace: e.NamespaceName,
		},
	}

	if _, err := controllerruntime.CreateOrUpdate(e.ctx, e.Client, ingress, func() error {
		if e.kls.TLSEnabled() {
			ingress.Spec.TLS = []v13.IngressTLS{
				{
					Hosts:      []string{e.kls.GetHost()},
					SecretName: e.cfg.SvcOpts.TLSSecretName,
				},
			}
		}

		pathType := v13.PathTypePrefix
		ingress.Labels = envLabels(e.kls)

		ingress.Spec.Rules = []v13.IngressRule{
			{
				Host: e.kls.GetHost(),
				IngressRuleValue: v13.IngressRuleValue{HTTP: &v13.HTTPIngressRuleValue{
					Paths: []v13.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: v13.IngressBackend{
								Service: &v13.IngressServiceBackend{
									Name: svc.GetName(),
									Port: v13.ServiceBackendPort{
										Name: svc.Spec.Ports[0].Name,
									},
								},
								Resource: nil,
							},
						},
					},
				}},
			},
		}

		return controllerruntime.SetControllerReference(e.kls, ingress, e.Client.Scheme())
	}); err != nil {
		return "", errors.Wrap(err, "error syncing ingress")
	}

	proto := "http://"
	if e.kls.TLSEnabled() {
		proto = "https://"
	}

	return proto + e.kls.GetHost(), nil
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
