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
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ServiceStore struct {
	cmClient  *rest.RESTClient
	clientset *kubernetes.Clientset
	log       logging.Logger
}

const (
	serviceK8sResource = "kuberlogicservices"
	readyStatus        = "Ready"
)

func (s *ServiceStore) GetService(name, namespace string, ctx context.Context) (*models.Service, bool, *ServiceError) {
	r := new(kuberlogicv1.KuberLogicService)

	err := s.cmClient.Get().
		Resource(serviceK8sResource).
		Namespace(namespace).
		Name(name).
		Do(ctx).
		Into(r)
	if err != nil && k8s.ErrNotFound(err) {
		s.log.Warnf("kuberlogic %s/%s not found: %s", namespace, name, err.Error())
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

func (s *ServiceStore) ListServices(ctx context.Context) ([]*models.Service, *ServiceError) {
	res := new(kuberlogicv1.KuberLogicServiceList)

	err := s.cmClient.Get().
		Resource(serviceK8sResource).
		Do(context.TODO()).
		Into(res)
	if err != nil {
		return nil, NewServiceError("error listing service", false, err)
	}
	s.log.Debugf("found %d kuberlogicservice objects: %v", len(res.Items), res)

	services := make([]*models.Service, len(res.Items))
	for i, r := range res.Items {
		service, err := s.kuberLogicToService(&r, ctx)
		if err != nil {
			return nil, NewServiceError("error converting service object", false, err)
		}
		services[i] = service
	}
	return services, nil
}

func (s *ServiceStore) CreateService(m *models.Service, ctx context.Context) (*models.Service, *ServiceError) {
	c, err := s.serviceToKuberLogic(m)
	if err != nil {
		return nil, NewServiceError("error converting service object", true, err)
	}

	_, found, _ := s.GetService(*m.Name, *m.Ns, ctx)
	if found {
		return nil, NewServiceError("service already exists", true, fmt.Errorf("service already exists"))
	}

	result := new(kuberlogicv1.KuberLogicService)
	err = s.cmClient.Post().
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

func (s *ServiceStore) UpdateService(m *models.Service, ctx context.Context) (*models.Service, *ServiceError) {
	// 1. see if exists
	currentC := new(kuberlogicv1.KuberLogicService)
	if err := s.cmClient.Get().
		Resource(serviceK8sResource).
		Namespace(*m.Ns).
		Name(*m.Name).
		Do(ctx).
		Into(currentC); err != nil && k8s.ErrNotFound(err) {
		return nil, NewServiceError("service not found", true, fmt.Errorf("service not found"))
	} else if err != nil {
		s.log.Errorf("service get error: %s", err.Error())
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

	s.log.Debugf("kuberlogic object result: %v", c)
	if err := s.cmClient.Put().
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

func (s *ServiceStore) DeleteService(m *models.Service, ctx context.Context) *ServiceError {
	_, f, getErr := s.GetService(*m.Name, *m.Ns, ctx)
	if !f {
		return NewServiceError("service not found", true, getErr.Err)
	}
	if getErr != nil {
		return getErr
	}

	err := s.cmClient.Delete().
		Resource(serviceK8sResource).
		Namespace(*m.Ns).
		Name(*m.Name).
		Do(ctx).
		Error()
	if err != nil {
		return NewServiceError("error deleting service", false, err)
	} else {
		return nil
	}
}

func (s *ServiceStore) GetServiceLogs(m *models.Service, instance string, lines int64, ctx context.Context) (string, *ServiceError) {
	m, f, errGet := s.GetService(*m.Name, *m.Ns, ctx)
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

func NewServiceStore(clientset *kubernetes.Clientset, cmClient *rest.RESTClient, logger logging.Logger) *ServiceStore {
	return &ServiceStore{
		cmClient:  cmClient,
		clientset: clientset,
		log:       logger,
	}
}

func (s *ServiceStore) NewServiceObject(name, namespace string) *models.Service {
	return &models.Service{Name: &name, Ns: &namespace}
}

func (s *ServiceStore) kuberLogicToService(c *kuberlogicv1.KuberLogicService, ctx context.Context) (*models.Service, error) {
	ret := new(models.Service)

	s.log.Debugf("converting kuberlogic %v to service", c)
	ret.Name = &c.Name
	ret.Ns = &c.Namespace
	ret.Type = &c.Spec.Type
	replicas := int64(c.Spec.Replicas - 1) // 1 - master
	ret.Replicas = &replicas
	ret.Masters = 1 // always equals 1

	ret.Status = c.Status.Status
	ret.CreatedAt = strfmt.DateTime(c.CreationTimestamp.Time)

	if ret.Status != readyStatus {
		s.log.Warnf("service %s/%s status is %s. not gathering more info", ret.Ns, ret.Name, ret.Status)
		return ret, nil
	}

	ret.Limits = new(models.Limits)
	if !c.Spec.Resources.Limits.Cpu().IsZero() {
		v, ok := c.Spec.Resources.Limits[v12.ResourceCPU]
		if ok {
			ret.Limits.CPU = strAsPointer(v.String())
		}
	}
	if !c.Spec.Resources.Limits.Memory().IsZero() {
		v, ok := c.Spec.Resources.Limits[v12.ResourceMemory]
		if ok {
			ret.Limits.Memory = strAsPointer(v.String())
		}
	}

	ret.Limits.VolumeSize = &c.Spec.VolumeSize

	ret.AdvancedConf = c.Spec.AdvancedConf

	ret.MaintenanceWindow = new(models.MaintenanceWindow)
	ret.MaintenanceWindow.Day = &c.Spec.MaintenanceWindow.Weekday
	ret.MaintenanceWindow.StartHour = int64AsPointer(int64(c.Spec.MaintenanceWindow.StartHour))

	instances, err := getServiceInstances(s.clientset, s.log, c, ctx)
	if err != nil {
		return ret, err
	}
	ret.Instances = instances

	intCon, err := getServiceInternalConnection(s.clientset, s.log, c)
	if err != nil {
		return ret, err
	}
	ret.InternalConnection = intCon

	extCon, err := getServiceExternalConnection(s.clientset, s.log, c)
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
			Namespace: *svc.Ns,
		},
	}

	if svc.Replicas != nil {
		// spec.Replicas equals svc.Replicas + svc.Master
		// svc.Master equals always 1 for pg/mysql
		c.Spec.Replicas = int32(*svc.Replicas + 1)
	}

	c.Spec.Type = *svc.Type
	// AP TODO: implememt version
	// c.Spec.Version = ""

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
