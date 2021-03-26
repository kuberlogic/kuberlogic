package store

import (
	"context"
	"fmt"
	"github.com/kuberlogic/operator/modules/apiserver/internal/generated/models"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"github.com/kuberlogic/operator/modules/apiserver/util/k8s"
	kuberlogicv1 "github.com/kuberlogic/operator/modules/operator/api/v1"
	util "github.com/kuberlogic/operator/modules/operator/service-operator/util/kuberlogic"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func mergeServices(m1 *models.Service, m2 *models.Service) (*models.Service, error) {
	if *m1.Name != *m1.Name || *m1.Ns != *m2.Ns {
		return nil, fmt.Errorf("name or namespace can't be changed")
	}

	if *m1.Type != *m2.Type {
		return nil, fmt.Errorf("type can't be changed")
	}

	if m2.AdvancedConf != nil {
		m1.AdvancedConf = m2.AdvancedConf
	}
	if m2.MaintenanceWindow != nil {
		m1.MaintenanceWindow = m2.MaintenanceWindow
	}

	if m2.Limits != nil {
		if m1.Limits == nil {
			m1.Limits = new(models.Limits)
		}

		if m2.Limits.Memory != nil {
			m1.Limits.Memory = m2.Limits.Memory
		}

		if m2.Limits.CPU != nil {
			m1.Limits.CPU = m2.Limits.CPU
		}

		if m2.Limits.VolumeSize != nil {
			m2Vol := resource.MustParse(*m2.Limits.VolumeSize)

			if m1.Limits.VolumeSize == nil {
				m1.Limits.VolumeSize = new(string)
			} else if m2Vol.Cmp(resource.MustParse(*m1.Limits.VolumeSize)) == -1 {
				return nil, fmt.Errorf("volume size can't be lowered")
			}

			m1.Limits.VolumeSize = m2.Limits.VolumeSize
		}
	}

	m1.Replicas = m2.Replicas

	return m1, nil
}

func getServiceExternalConnection(c *kubernetes.Clientset, log logging.Logger, cm *kuberlogicv1.KuberLogicService) (*models.ServiceExternalConnection, error) {
	svc := new(models.ServiceExternalConnection)

	masterSvc, replicaSvc, err := util.GetClusterServices(cm)
	if err != nil {
		return svc, err
	}
	log.Debugw("services", "master", masterSvc, "replica", replicaSvc)

	port, err := util.GetClusterServicePort(cm)
	if err != nil {
		return svc, err
	}

	masterExt, _, err := k8s.GetServiceExternalIP(c, log, masterSvc, cm.Namespace)
	if err != nil {
		return svc, err
	}
	replicaExt, _, err := k8s.GetServiceExternalIP(c, log, replicaSvc, cm.Namespace)
	if err != nil {
		return svc, err
	}

	user, password, err := getServiceCredentials(c, log, cm)
	if err != nil {
		return nil, err
	}

	svc.Master = &models.Connection{
		Cert:     "",
		Host:     masterExt,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	svc.Replica = &models.Connection{
		Cert:     "",
		Host:     replicaExt,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	return svc, nil
}

func getServiceInternalConnection(c *kubernetes.Clientset, log logging.Logger, cm *kuberlogicv1.KuberLogicService) (*models.ServiceInternalConnection, error) {
	svc := new(models.ServiceInternalConnection)

	masterSvc, replicaSvc, err := util.GetClusterServices(cm)
	if err != nil {
		return svc, err
	}

	port, err := util.GetClusterServicePort(cm)
	if err != nil {
		return svc, err
	}

	user, password, err := getServiceCredentials(c, log, cm)
	if err != nil {
		return nil, err
	}

	svc.Master = &models.Connection{
		Cert:     "",
		Host:     masterSvc,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	svc.Replica = &models.Connection{
		Cert:     "",
		Host:     replicaSvc,
		Password: password,
		Port:     int64(port),
		SslMode:  "",
		User:     user,
	}
	return svc, nil
}

func getServiceCredentials(c *kubernetes.Clientset, log logging.Logger, cm *kuberlogicv1.KuberLogicService) (user, password string, err error) {
	user, passwordField, secretName, err := util.GetClusterCredentialsInfo(cm)
	if err != nil {
		err = fmt.Errorf("Error getting connecion credentials: %s", err.Error())
		return
	}

	user = kuberlogicv1.DefaultUser
	log.Debugw("trying to get credentials for username",
		"user", user, "secret", user, "password", passwordField)
	password, err = k8s.GetSecretFieldDecoded(c, log, secretName, cm.Namespace, passwordField)
	if err != nil {
		return user, password, err
	}
	return
}

func getServiceInstances(c *kubernetes.Clientset, log logging.Logger, cm *kuberlogicv1.KuberLogicService, ctx context.Context) ([]*models.ServiceInstance, error) {
	masterPods, replicaPods, err := getServicePods(c, log, cm, ctx)
	if err != nil {
		return nil, err
	}

	var instances []*models.ServiceInstance
	instances = append(instances, podsToServiceInstances(masterPods, "master")...)
	instances = append(instances, podsToServiceInstances(replicaPods, "replica")...)

	return instances, nil
}

func podsToServiceInstances(pods *corev1.PodList, role string) (instances []*models.ServiceInstance) {
	for _, p := range pods.Items {
		instances = append(instances, &models.ServiceInstance{
			Name:   p.Name,
			Role:   role,
			Status: podStatusToServiceInstanceStatus(p),
		})
	}
	return instances
}

func podStatusToServiceInstanceStatus(p corev1.Pod) *models.ServiceInstanceStatus {
	s := &models.ServiceInstanceStatus{}
	switch p.Status.Phase {
	case "Pending":
		s.Status = "Pending"
		break
	case "Running":
		s.Status = "Running"
		break
	default:
		s.Status = "Failed"
	}
	return s
}

func getServicePods(c *kubernetes.Clientset, log logging.Logger, cm *kuberlogicv1.KuberLogicService, ctx context.Context) (masterPods *corev1.PodList, replicaPods *corev1.PodList, err error) {
	masterPodSelector, replicaPodSelector, err := util.GetClusterPodLabels(cm)
	if err != nil {
		return
	}

	podListOpts := metav1.ListOptions{
		LabelSelector: k8s.MapToStrSelector(masterPodSelector),
	}
	masterPods, err = c.CoreV1().Pods(cm.Namespace).List(ctx, podListOpts)
	log.Debugw("master pods details",
		"master pods", &masterPods, "pod list options", podListOpts)
	if err != nil {
		return
	}

	podListOpts.LabelSelector = k8s.MapToStrSelector(replicaPodSelector)
	replicaPods, err = c.CoreV1().Pods(cm.Namespace).List(ctx, podListOpts)
	if err != nil {
		return
	}
	log.Debugw("replica pods details", "replica pods", &replicaPods, "pod list options", podListOpts)
	return
}

func getServiceInstanceLogs(c *kubernetes.Clientset, cm *kuberlogicv1.KuberLogicService, log logging.Logger, ctx context.Context, instance string, lines int64) (logs string, found bool, err error) {
	found = true

	instances, err := getServiceInstances(c, log, cm, ctx)
	if err != nil {
		return logs, false, fmt.Errorf("error getting service instances: %s", err.Error())
	}

	for _, p := range append(instances) {
		if p.Name == instance {
			container, _ := util.GetClusterMainContainer(cm)
			logs, err = k8s.GetPodLogs(c, log, p.Name, container, cm.Namespace, lines)
			if err != nil {
				return
			}
			return
		}
	}
	return logs, false, nil
}

func int64AsPointer(x int64) *int64 {
	return &x
}

func strAsPointer(x string) *string {
	return &x
}
