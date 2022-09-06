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
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type EnvironmentManager struct {
	client.Client

	NamespaceName string

	kls *kuberlogiccomv1alpha1.KuberLogicService
	cfg *config.Config
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=namespaces;services;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=resourcequotas;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=secrets;,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods;,verbs=deletecollection
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicebackupschedules,verbs=get;list;watch;create;update;patch;delete

// SetupEnv checks if KLS environment is present and creates it if it is not
func (e *EnvironmentManager) SetupEnv(ctx context.Context) error {
	// Namespace ns will contain all managed resources
	ns := &v1.Namespace{
		ObjectMeta: v12.ObjectMeta{
			Name:   e.NamespaceName,
			Labels: envLabels(e.kls),
		},
	}
	if _, err := controllerruntime.CreateOrUpdate(ctx, e.Client, ns, func() error {
		return controllerruntime.SetControllerReference(e.kls, ns, e.Scheme())
	}); err != nil {
		return errors.Wrap(err, "error setting up kls namespace")
	}

	// NetworkPolicy netpol prevents all cross namespace network communication for service pods
	netpol := &v13.NetworkPolicy{
		ObjectMeta: v12.ObjectMeta{
			Name:      "namespace-only-egress",
			Namespace: ns.Name,
			Labels:    envLabels(e.kls),
		},
	}
	if _, err := controllerruntime.CreateOrUpdate(ctx, e.Client, netpol, func() error {
		netpol.Spec = v13.NetworkPolicySpec{
			PolicyTypes: []v13.PolicyType{
				v13.PolicyTypeEgress,
			},
			Egress: []v13.NetworkPolicyEgressRule{
				{
					To: []v13.NetworkPolicyPeer{
						{
							NamespaceSelector: &v12.LabelSelector{
								MatchLabels: envLabels(e.kls),
							},
						},
					},
				},
			},
		}
		return controllerruntime.SetControllerReference(e.kls, netpol, e.Scheme())
	}); err != nil {
		return errors.Wrap(err, "error setting up kls networkpolicy")
	}

	// Sync TLS secret when defined in config
	tlsSecret := &v1.Secret{
		ObjectMeta: v12.ObjectMeta{
			Name:      e.cfg.SvcOpts.TLSSecretName,
			Namespace: ns.GetName(),
		},
		Type: v1.SecretTypeTLS,
	}

	if e.cfg.SvcOpts.TLSSecretName != "" {
		if _, err := controllerruntime.CreateOrUpdate(ctx, e.Client, tlsSecret, func() error {
			srcSecret := &v1.Secret{
				ObjectMeta: v12.ObjectMeta{
					Name:      e.cfg.SvcOpts.TLSSecretName,
					Namespace: e.cfg.Namespace,
				},
			}
			if err := e.Get(ctx, client.ObjectKeyFromObject(srcSecret), srcSecret); err != nil {
				return errors.Wrap(err, "error getting source TLS secret")
			}
			tlsSecret.Data = srcSecret.Data
			return controllerruntime.SetControllerReference(e.kls, tlsSecret, e.Scheme())
		}); err != nil {
			return errors.Wrap(err, "error syncing TLS Secret")
		}
	}

	klbs := &kuberlogiccomv1alpha1.KuberlogicServiceBackupSchedule{}
	klbs.SetName(e.kls.GetName())
	klbs.SetNamespace(e.cfg.Namespace)
	if e.kls.Spec.BackupSchedule != "" {
		if _, err := controllerruntime.CreateOrUpdate(ctx, e.Client, klbs, func() error {
			klbs.Spec.KuberlogicServiceName = e.kls.GetName()
			klbs.Spec.Schedule = e.kls.Spec.BackupSchedule
			return controllerruntime.SetControllerReference(e.kls, klbs, e.Scheme())
		}); err != nil {
			return errors.Wrap(err, "failed to enabled backup schedule")
		}
	} else {
		if err := e.Delete(ctx, klbs); err != nil && !errors2.IsNotFound(err) {
			return errors.Wrap(err, "failed to disable scheduled backups")
		}
	}

	// set namespace status field
	e.kls.Status.Namespace = ns.GetName()
	return nil
}

// PauseService deletes all pods in a service namespace.
// In addition to this resourceQuota in SetupEnv sets the hard limit of non-exited pods to 0.
func (e *EnvironmentManager) PauseService(ctx context.Context) error {
	return e.deleteAllServicePods(ctx)
}

func (e *EnvironmentManager) ResumeService(ctx context.Context) error {
	return e.deleteAllServicePods(ctx)
}

func (e *EnvironmentManager) deleteAllServicePods(ctx context.Context) error {
	if err := e.DeleteAllOf(ctx, &v1.Pod{}, &client.DeleteAllOfOptions{
		ListOptions: client.ListOptions{
			Namespace: e.NamespaceName,
		},
		DeleteOptions: client.DeleteOptions{},
	}); err != nil {
		return errors.Wrap(err, "error deleting all pods in namespace")
	}
	return nil
}

func New(c client.Client, kls *kuberlogiccomv1alpha1.KuberLogicService, cfg *config.Config) *EnvironmentManager {
	return &EnvironmentManager{
		Client: c,

		NamespaceName: kls.GetName(),

		cfg: cfg,
		kls: kls,
	}
}

func envLabels(kls *kuberlogiccomv1alpha1.KuberLogicService) map[string]string {
	return map[string]string{
		"kuberlogic.com/env": kls.Name,
	}
}
