/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package v1alpha1

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kuberlogicservicebackuplog = logf.Log.WithName("kuberlogicservicebackup-resource")

var backupsEnabled bool

var (
	backupsDisabledError = errors.New("backups disabled in config")
)

func (r *KuberlogicServiceBackup) SetupWebhookWithManager(mgr ctrl.Manager, cfgBackupsEnabled bool) error {
	backupsEnabled = cfgBackupsEnabled
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kuberlogic-com-v1alpha1-kuberlogicservicebackup,mutating=true,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservicebackups,verbs=create;update,versions=v1alpha1,name=mkuberlogicservicebackup.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &KuberlogicServiceBackup{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *KuberlogicServiceBackup) Default() {
	kuberlogicservicebackuplog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-kuberlogic-com-v1alpha1-kuberlogicservicebackup,mutating=false,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservicebackups,verbs=create;update;delete,versions=v1alpha1,name=vkuberlogicservicebackup.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &KuberlogicServiceBackup{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberlogicServiceBackup) ValidateCreate() error {
	kuberlogicservicebackuplog.Info("validate create", "name", r.Name)
	if !backupsEnabled {
		return backupsDisabledError
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *KuberlogicServiceBackup) ValidateUpdate(old runtime.Object) error {
	kuberlogicservicebackuplog.Info("validate update", "name", r.Name)
	if !backupsEnabled {
		return backupsDisabledError
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *KuberlogicServiceBackup) ValidateDelete() error {
	kuberlogicservicebackuplog.Info("validate delete", "name", r.Name)
	return nil
}
