/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package controllers

import (
	"context"
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/cfg"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	kuberlogiccomv1alpha1 "github.com/kuberlogic/kuberlogic/modules/dynamic-operator/api/v1alpha1"
)

// KuberlogicServiceBackupScheduleReconciler reconciles a KuberlogicServiceBackupSchedule object
type KuberlogicServiceBackupScheduleReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Cfg    *cfg.Config
}

//+kubebuilder:rbac:groups=kuberlogic.com.kuberlogic.com,resources=kuberlogicservicebackupschedules,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kuberlogic.com.kuberlogic.com,resources=kuberlogicservicebackupschedules/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kuberlogic.com.kuberlogic.com,resources=kuberlogicservicebackupschedules/finalizers,verbs=update

//+kubebuilder:rbac:groups=kuberlogic.com,resources=kuberlogicservicebackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="rbac.authorization.k8s.io",resources=rolebindings,verbs=get;list;watch;create;update;patch;delete

func (r *KuberlogicServiceBackupScheduleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx).WithValues("name", req.String())

	klbs := &kuberlogiccomv1alpha1.KuberlogicServiceBackupSchedule{}
	if err := r.Get(ctx, req.NamespacedName, klbs); err != nil {
		if k8serrors.IsNotFound(err) {
			l.Info("object not found", "key", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		l.Error(err, "Failed to get KuberlogicServiceBackupSchedule")
		return ctrl.Result{}, err
	}

	kls := &kuberlogiccomv1alpha1.KuberLogicService{}
	kls.SetName(klbs.Spec.KuberlogicServiceName)

	if err := r.Get(ctx, client.ObjectKeyFromObject(kls), kls); k8serrors.IsNotFound(err) {
		l.Error(err, "service not found")
		return ctrl.Result{}, err
	} else if err != nil {
		l.Error(err, "error getting service", "name", kls.GetName())
		return ctrl.Result{}, err
	}

	periodicBackupCJ := &batchv1.CronJob{}
	periodicBackupCJ.SetName(klbs.GetName())
	periodicBackupCJ.SetNamespace(klbs.GetNamespace())
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, periodicBackupCJ, func() error {
		periodicBackupCJ.Spec.Schedule = klbs.Spec.Schedule
		periodicBackupCJ.Spec.ConcurrencyPolicy = batchv1.ForbidConcurrent
		jobHistoryLimits := int32(1)
		periodicBackupCJ.Spec.SuccessfulJobsHistoryLimit = &jobHistoryLimits
		periodicBackupCJ.Spec.FailedJobsHistoryLimit = &jobHistoryLimits

		backoffLimit := int32(2)
		periodicBackupCJ.Spec.JobTemplate.Spec.BackoffLimit = &backoffLimit
		activeDeadlineSeconds := int64(15)
		periodicBackupCJ.Spec.JobTemplate.Spec.ActiveDeadlineSeconds = &activeDeadlineSeconds

		periodicBackupCJ.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyOnFailure
		periodicBackupCJ.Spec.JobTemplate.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  "klb-create",
				Image: "bitnami/kubectl:1.23.8",
				Command: []string{
					"/bin/sh",
				},
				Args: []string{
					"-c",
					fmt.Sprintf("echo '{\"apiVersion\":\"kuberlogic.com/v1alpha1\",\"kind\":\"KuberlogicServiceBackup\",\"metadata\":{\"name\":\"%s\", \"labels\":{\"kls-id\": \"%s\"}},\"spec\":{\"kuberlogicServiceName\":\"%s\"}}' | kubectl apply -f -", klbs.GetName(), klbs.Spec.KuberlogicServiceName, klbs.Spec.KuberlogicServiceName),
				},
			},
		}
		periodicBackupCJ.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName = r.Cfg.ServiceAccount

		return ctrl.SetControllerReference(klbs, periodicBackupCJ, r.Scheme)
	}); err != nil {
		l.Error(err, "failed to setup scheduled backup cronjob")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KuberlogicServiceBackupScheduleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kuberlogiccomv1alpha1.KuberlogicServiceBackupSchedule{}).
		Owns(&batchv1.CronJob{}).
		Complete(r)
}
