/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// KuberLogicServiceReconciler reconciles a KuberLogicService object
type KuberLogicServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func HandlePanic(log logr.Logger) {
	if err := recover(); err != nil {
		log.Error(errors.New("handle panic"), fmt.Sprintf("%v", err))
		result := sentry.Flush(5 * time.Second)
		if !result {
			time.Sleep(5 * time.Second)
		}
		panic(err)
	}
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservices/finalizers,verbs=update

func (r *KuberLogicServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx).WithValues("kuberlogicservicetype", req.String())
	log.Info("Reconciliation started")
	defer HandlePanic(log)

	// Fetch the KuberLogicServices instance
	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	err := r.Get(ctx, req.NamespacedName, kls)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info(req.Namespace, req.Name, " is absent")

			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get KuberLogicService")
		return ctrl.Result{}, err
	}

	spec := make(map[string]interface{}, 0)
	if err := json.Unmarshal(kls.Spec.Raw, &spec); err != nil {
		log.Error(err, "error unmarshaling spec")
		return ctrl.Result{}, err
	}

	klst := &kuberlogiccomv1alpha1.KuberLogicServiceType{}
	err = r.Get(ctx, types.NamespacedName{
		Name:      kls.GetServiceType(spec),
		Namespace: req.Namespace,
	}, klst)
	if err != nil {
		log.Error(err, "Failed to get KuberLogicServiceType")
		return ctrl.Result{}, err
	}

	defaultSpec := make(map[string]interface{}, 0)
	if err := json.Unmarshal(klst.Spec.DefaultSpec.Raw, &defaultSpec); err != nil {
		log.Error(err, "error unmarshaling defaultSpec")
		return ctrl.Result{}, err
	}

	// define a basic svc object to work with it
	svc := &unstructured.Unstructured{}
	svc.SetGroupVersionKind(klst.ServiceGVK())
	svc.SetName(kls.Name)
	svc.SetNamespace(kls.Namespace)
	if err := unstructured.SetNestedField(svc.UnstructuredContent(), defaultSpec, "spec"); err != nil {
		log.Error(err, "error setting default spec", "defaultSpec", defaultSpec)
		return ctrl.Result{}, err
	}

	log.Info("debug", "service", svc.UnstructuredContent())
	if err := r.Client.Get(ctx, req.NamespacedName, svc); k8serrors.IsNotFound(err) {
		log.Info("creating new service", "type", kls.GetServiceType(spec))

		err = r.SetFields(spec, svc, klst, log)
		if err != nil {
			log.Error(err, "error parsing fields")
			return ctrl.Result{}, err
		}

		if err := ctrl.SetControllerReference(kls, svc, r.Scheme); err != nil {
			log.Error(err, "error setting controller reference")
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, svc); err != nil {
			log.Error(err, "error creaing service object")
			return ctrl.Result{}, err
		}
	} else if err != nil {
		log.Error(err, "failed to get service object")
		return ctrl.Result{}, err
	}
	log.Info("service object exists")
	log.Info("syncing kuberlogicservice parameters")

	err = r.SetFields(spec, svc, klst, log)
	if err != nil {
		log.Error(err, "error parsing fields")
		return ctrl.Result{}, err
	}

	log.Info("updating the service", "svc", svc.UnstructuredContent())
	if err := r.Update(ctx, svc); err != nil {
		log.Error(err, "error updating service object")
		return ctrl.Result{}, err
	}

	log.Info("syncing status", "object", svc.UnstructuredContent())

	kls.MarkNotReady("Ready condition not found")
	conditions, found, err := unstructured.NestedSlice(svc.UnstructuredContent(), strings.Split(klst.Spec.StatusRef.Conditions.Path, ".")...)
	if err != nil {
		log.Error(err, "conditions is not found in service object")
		return ctrl.Result{}, err
	}
	if !found {
		log.Info("Ready condition is not found in service object")
		kls.MarkNotReady("ReadyConditionNotFound")
	} else {
		for _, c := range conditions {
			cond := c.(map[string]interface{})
			if cond["type"].(string) == klst.Spec.StatusRef.Conditions.ReadyCondition {
				if cond["status"] == klst.Spec.StatusRef.Conditions.ReadyValue {
					kls.MarkReady("ReadyConditionMet")
				} else {
					kls.MarkNotReady("ReadyConditionNotMet")
				}
			}
		}
	}
	if err := r.Status().Update(ctx, kls); err != nil {
		log.Error(err, "error syncing status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func setField(svc *unstructured.Unstructured, value interface{}, path string) error {
	pathSliced := strings.Split(path, ".")
	return unstructured.SetNestedField(svc.UnstructuredContent(), value, pathSliced...)
}

func (r *KuberLogicServiceReconciler) SetFields(
	spec map[string]interface{},
	svc *unstructured.Unstructured,
	klst *kuberlogiccomv1alpha1.KuberLogicServiceType,
	log logr.Logger,
) error {
	for k, typeValue := range klst.Spec.SpecRef {
		value, _ := spec[k]
		if err := setField(svc, value, typeValue.Path); err != nil {
			log.Error(err, "error setting value")
			return err
		}
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KuberLogicServiceReconciler) SetupWithManager(mgr ctrl.Manager) (controller.Controller, error) {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogiccomv1alpha1.KuberLogicService{}).
		Build(r)
}
