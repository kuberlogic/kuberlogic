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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// KuberLogicServiceTypeReconciler reconciles a KuberLogicServiceType object
type KuberLogicServiceTypeReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ServiceController controller.Controller
}

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicetypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicetypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicetypes/finalizers,verbs=update

func (r *KuberLogicServiceTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logger.FromContext(ctx)

	klst := new(kuberlogiccomv1alpha1.KuberLogicServiceType)
	err := r.Client.Get(ctx, req.NamespacedName, klst)
	if errors.IsNotFound(err) {
		log.Info("object not found")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "error retrieving object")
		return ctrl.Result{}, err
	}

	gvk := schema.GroupVersionKind{Group: klst.Spec.Api.Group, Version: klst.Spec.Api.Version, Kind: klst.Spec.Api.Kind}
	obj := &unstructured.Unstructured{}
	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	obj.SetGroupVersionKind(gvk)

	err = r.ServiceController.Watch(&source.Kind{
		Type: obj,
	}, &handler.EnqueueRequestForOwner{
		OwnerType:    kls,
		IsController: true,
	})
	if err != nil {
		log.Error(err, "error configuring watch the resource", "type", obj)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KuberLogicServiceTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogiccomv1alpha1.KuberLogicServiceType{}).
		Complete(r)
}
