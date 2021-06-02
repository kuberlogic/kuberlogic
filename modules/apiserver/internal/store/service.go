package store

import (
	"context"
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	v12 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	serviceK8sResource = "kuberlogicservices"
	readyStatus        = "Ready"
)

func (s *ServiceStore) GetService(name, namespace string, ctx context.Context) (*models.Service, bool, *ServiceError) {
	r := new(kuberlogicv1.KuberLogicService)

	err := s.restClient.Get().
		Resource(serviceK8sResource).
		Namespace(namespace).
		Name(name).
		Do(ctx).
		Into(r)
	if err != nil && k8s.ErrNotFound(err) {
		s.log.Warnw("kuberlogic service not found",
			"namespace", namespace, "name", name, "error", err)
		return nil, false, NewServiceError("service not found", true, fmt.Errorf("service not found"))
	} else if err != nil {
		return nil, true, NewServiceError("error getting service", false, err)
	}

	ret, err := s.kuberLogicToService(r, ctx)
	if err != nil {
		return nil, false, NewServiceError("error converting service object", false, err)
	}

	return ret, true, nil
}

func (s *ServiceStore) ListServices(p *models.Principal, ctx context.Context) ([]*models.Service, *ServiceError) {
	res := new(kuberlogicv1.KuberLogicServiceList)

	err := s.restClient.Get().
		Resource(serviceK8sResource).
		Namespace(p.Namespace).
		Do(context.TODO()).
		Into(res)
	if err != nil {
		return nil, NewServiceError("error listing service", false, err)
	}
	s.log.Debugw("found kuberlogicservice objects", "length", len(res.Items), "objects", res)

	var services []*models.Service
	for _, r := range res.Items {
		service, err := s.kuberLogicToService(&r, ctx)
		if err != nil {
			return nil, NewServiceError("error converting service object", false, err)
		}
		services = append(services, service)
	}
	return services, nil
}

func (s *ServiceStore) CreateService(m *models.Service, p *models.Principal, ctx context.Context) (*models.Service, *ServiceError) {
	s.log.Debugw("create service request", "service", m, "principal", p)
	// ensure that namespace exists before we create a service
	if err := s.ensureNamespace(p.Namespace, ctx); err != nil {
		return nil, NewServiceError("error creating service namespace", false, err)
	}
	c, err := s.serviceToKuberLogic(m)
	if err != nil {
		return nil, NewServiceError("error converting service object", true, err)
	}
	if err := c.SetAlertEmail(p.Email); err != nil {
		return nil, NewServiceError("error setting email for monitoring notifications", true, err)
	}

	_, found, _ := s.GetService(*m.Name, m.Ns, ctx)
	if found {
		return nil, NewServiceError("service already exists", true, fmt.Errorf("service already exists"))
	}

	result := new(kuberlogicv1.KuberLogicService)
	err = s.restClient.Post().
		Resource(serviceK8sResource).
		Namespace(c.Namespace).
		Name(c.Name).
		Body(c).
		Do(ctx).
		Into(result)
	if err != nil {
		return nil, NewServiceError("error creating service", false, err)
	}
	svc, err := s.kuberLogicToService(result, ctx)
	if err != nil {
		return nil, NewServiceError("error getting newly created service", false, err)
	}
	return svc, nil
}

func (s *ServiceStore) UpdateService(m *models.Service, p *models.Principal, ctx context.Context) (*models.Service, *ServiceError) {
	// 1. see if exists
	currentC := new(kuberlogicv1.KuberLogicService)
	if err := s.restClient.Get().
		Resource(serviceK8sResource).
		Namespace(p.Namespace).
		Name(*m.Name).
		Do(ctx).
		Into(currentC); err != nil && k8s.ErrNotFound(err) {
		return nil, NewServiceError("service not found", true, fmt.Errorf("service not found"))
	} else if err != nil {
		s.log.Errorw("service get error", "error", err)
		return nil, NewServiceError("error getting service", false, err)
	}

	current, err := s.kuberLogicToService(currentC, ctx)
	if err != nil {
		return nil, NewServiceError("error converting service object", false, err)
	}
	wanted, errMerge := mergeServices(current, m)
	if errMerge != nil {
		return nil, NewServiceError(fmt.Sprintf("error changing service: %s", errMerge.Error()), true, errMerge)
	}

	c, errConvert := s.serviceToKuberLogic(wanted)
	if errConvert != nil {
		return nil, NewServiceError("error converting service object", false, errConvert)
	}
	c.ResourceVersion = currentC.ResourceVersion

	s.log.Debugw("kuberlogic object result", "body", c)
	if err := s.restClient.Put().
		Resource(serviceK8sResource).
		Name(c.Name).
		Namespace(c.Namespace).
		Body(c).
		Do(ctx).
		Error(); err != nil {
		return nil, NewServiceError("error updating service", false, err)
	}

	return wanted, nil
}

func (s *ServiceStore) DeleteService(m *models.Service, p *models.Principal, ctx context.Context) *ServiceError {
	_, f, getErr := s.GetService(*m.Name, p.Namespace, ctx)
	if !f {
		return NewServiceError("service not found", true, getErr.Err)
	}
	if getErr != nil {
		return getErr
	}

	err := s.restClient.Delete().
		Resource(serviceK8sResource).
		Namespace(p.Namespace).
		Name(*m.Name).
		Do(ctx).
		Error()
	if err != nil {
		return NewServiceError("error deleting service", false, err)
	}

	return nil
}

func (s *ServiceStore) GetServiceLogs(m *models.Service, instance string, lines int64, ctx context.Context) (string, *ServiceError) {
	m, f, errGet := s.GetService(*m.Name, m.Ns, ctx)
	if errGet != nil {
		return "", errGet
	}
	if !f {
		return "", NewServiceError("service not found", true, fmt.Errorf("service not found"))
	}

	c, err := s.serviceToKuberLogic(m)
	if err != nil {
		return "", NewServiceError("error converting service object", false, err)
	}

	logs, found, errLogs := getServiceInstanceLogs(s.clientset, c, s.log, ctx, instance, lines)
	if errLogs != nil {
		return "", NewServiceError("error getting service logs", false, errLogs)
	}
	if !found {
		return "", NewServiceError("service instance not found", true, fmt.Errorf("service instance not found"))
	}
	return logs, nil
}

func NewServiceStore(clientset *kubernetes.Clientset, restClient *rest.RESTClient, logger logging.Logger) *ServiceStore {
	return &ServiceStore{
		restClient: restClient,
		clientset:  clientset,
		log:        logger,
	}
}

func (s *ServiceStore) NewServiceObject(name, namespace string) *models.Service {
	return &models.Service{Name: &name, Ns: namespace}
}

func (s *ServiceStore) kuberLogicToService(kls *kuberlogicv1.KuberLogicService, ctx context.Context) (*models.Service, error) {
	ret := new(models.Service)
	s.log.Debugw("converting kuberlogic to service", "kuberlogic service", kls)
	ret.Name = strAsPointer(kls.Name)
	ret.Ns = kls.Namespace
	ret.Type = strAsPointer(kls.Spec.Type)
	ret.Replicas = int64AsPointer(int64(kls.Spec.Replicas - 1)) // 1 - master
	ret.Masters = 1                                             // always equals 1

	ready, status := kls.IsReady()
	ret.Status = status
	ret.CreatedAt = strfmt.DateTime(kls.CreationTimestamp.Time.UTC())

	if !ready {
		s.log.Warnw(fmt.Sprintf("service status is not equal %s. not gathering more info", readyStatus),
			"namespace", ret.Ns, "name", ret.Name, "status", ret.Status)
		return ret, nil
	}

	ret.Limits = new(models.Limits)
	if !kls.Spec.Resources.Limits.Cpu().IsZero() {
		v, ok := kls.Spec.Resources.Limits[v12.ResourceCPU]
		if ok {
			ret.Limits.CPU = strAsPointer(v.String())
		}
	}
	if !kls.Spec.Resources.Limits.Memory().IsZero() {
		v, ok := kls.Spec.Resources.Limits[v12.ResourceMemory]
		if ok {
			ret.Limits.Memory = strAsPointer(v.String())
		}
	}

	ret.Limits.VolumeSize = &kls.Spec.VolumeSize

	ret.AdvancedConf = kls.Spec.AdvancedConf

	ret.MaintenanceWindow = new(models.MaintenanceWindow)
	ret.MaintenanceWindow.Day = strAsPointer(kls.Spec.MaintenanceWindow.Weekday)
	ret.MaintenanceWindow.StartHour = int64AsPointer(int64(kls.Spec.MaintenanceWindow.StartHour))

	instances, err := getServiceInstances(s.clientset, s.log, kls, ctx)
	if err != nil {
		return ret, err
	}
	ret.Instances = instances

	intCon, err := getServiceInternalConnection(s.clientset, s.log, kls)
	if err != nil {
		return ret, err
	}
	ret.InternalConnection = intCon

	extCon, err := getServiceExternalConnection(s.clientset, s.log, kls)
	if err != nil {
		return ret, err
	}
	ret.ExternalConnection = extCon

	return ret, nil
}

func (s *ServiceStore) serviceToKuberLogic(svc *models.Service) (*kuberlogicv1.KuberLogicService, error) {
	c := &kuberlogicv1.KuberLogicService{
		ObjectMeta: v1.ObjectMeta{
			Name:      *svc.Name,
			Namespace: svc.Ns,
		},
	}

	if svc.Replicas != nil {
		// spec.Replicas equals svc.Replicas + svc.Master
		// svc.Master equals always 1 for pg/mysql
		c.Spec.Replicas = int32(*svc.Replicas + 1)
	}

	c.Spec.Type = *svc.Type
	if svc.Version != "" {
		c.Spec.Version = svc.Version
	}

	if svc.Limits != nil {
		c.Spec.Resources.Limits = make(v12.ResourceList)
		c.Spec.Resources.Requests = make(v12.ResourceList)

		cpu := svc.Limits.CPU
		if cpu != nil {
			// amount of resources and limits could be different
			// for using the same values need to use the same defaults in the operator's scope
			c.Spec.Resources.Limits[v12.ResourceCPU] = resource.MustParse(*svc.Limits.CPU)
		}

		mem := svc.Limits.Memory
		if mem != nil {
			// amount of resources and limits could be different
			// for using the same values need to use the same defaults in the operator's scope
			c.Spec.Resources.Limits[v12.ResourceMemory] = resource.MustParse(*svc.Limits.Memory)
		}

		if svc.Limits.VolumeSize != nil {
			c.Spec.VolumeSize = *svc.Limits.VolumeSize
		}
	}

	if svc.MaintenanceWindow != nil {
		c.Spec.MaintenanceWindow = kuberlogicv1.MaintenanceWindow{}
		c.Spec.MaintenanceWindow.Weekday = *svc.MaintenanceWindow.Day
		c.Spec.MaintenanceWindow.StartHour = int(*svc.MaintenanceWindow.StartHour)
	}

	if svc.AdvancedConf != nil {
		c.Spec.AdvancedConf = svc.AdvancedConf
	}

	return c, nil
}

func (s *ServiceStore) ensureNamespace(namespace string, ctx context.Context) error {
	ns := &v12.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: namespace,
			Annotations: map[string]string{
				"app.kubernetes.io/managed-by": "kuberlogic-apiserver",
			},
		},
	}
	_, err := s.clientset.CoreV1().Namespaces().Create(ctx, ns, v1.CreateOptions{})
	if err != nil && errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
