package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var kuberlogicSerivceLog = logf.Log.WithName("kuberlogic-service-resource")

func (kls *KuberLogicService) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(kls).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-kuberlogic-com-v1-kuberlogicservice,mutating=true,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update,versions=v1,name=mkuberlogicservice.kuberlogic.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &KuberLogicService{}

func (kls *KuberLogicService) Default() {
	kuberlogicSerivceLog.Info("default", "name", kls.Name)

}

//+kubebuilder:webhook:path=/validate-kuberlogic-com-v1-kuberlogicservice,mutating=false,failurePolicy=fail,sideEffects=None,groups=kuberlogic.com,resources=kuberlogicservices,verbs=create;update,versions=v1,name=vkuberlogicservice.kuberlogic.com,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &KuberLogicService{}

func (kls *KuberLogicService) ValidateCreate() error {
	kuberlogicSerivceLog.Info("validate create", "name", kls.Name)

	return nil
}

func (kls *KuberLogicService) ValidateUpdate(old runtime.Object) error {
	kuberlogicSerivceLog.Info("validate update", "name", kls.Name)

	return nil
}

func (kls *KuberLogicService) ValidateDelete() error {
	kuberlogicSerivceLog.Info("validate delete", "name", kls.Name)

	return nil
}
