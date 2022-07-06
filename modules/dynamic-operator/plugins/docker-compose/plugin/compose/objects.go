package compose

import (
	"bytes"
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"text/template"
)

var (
	serviceAccountGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ServiceAccount"}
	serviceGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Service",
	}
	pvcGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "PersistentVolumeClaim"}
	deploymentGVK = schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment"}
)

var (
	errUnknownObject = errors.New("unknown object kind")
)

type ComposeModel struct {
	composeProject *types.Project
	logger         hclog.Logger

	serviceaccount        *corev1.ServiceAccount
	service               *corev1.Service
	persistentvolumeclaim *corev1.PersistentVolumeClaim
	deployment            *appsv1.Deployment
}

// Reconcile method updates current request object to their required parameters
func (c *ComposeModel) Reconcile(req *commons.PluginRequest) ([]map[schema.GroupVersionKind]client.Object, error) {
	c.logger.Debug("Reconcile")

	existingObjects := req.GetObjects()
	if err := c.fromCluster(existingObjects); err != nil {
		return nil, errors.Wrap(err, "error marshaling cluster objects")
	}

	if err := c.setObjects(req); err != nil {
		return nil, errors.Wrap(err, "error updating service objects")
	}
	return c.objectsWithGVK(), nil
}

// Ready checks if compose application is running
func (c *ComposeModel) Ready(req *commons.PluginRequest) (bool, error) {
	c.logger.Debug("Ready")

	existingObjects := req.GetObjects()
	if err := c.fromCluster(existingObjects); err != nil {
		return false, errors.Wrap(err, "error marshaling cluster objects")
	}
	return c.isReady(), nil
}

// Types returns list of empty objects with their GVK
func (c *ComposeModel) Types() []map[schema.GroupVersionKind]client.Object {
	c.logger.Debug("Type")
	return []map[schema.GroupVersionKind]client.Object{
		{
			serviceGVK: &corev1.Service{},
		},
		{
			serviceAccountGVK: &corev1.ServiceAccount{},
		},
		{
			pvcGVK: &corev1.PersistentVolumeClaim{},
		},
		{
			deploymentGVK: &appsv1.Deployment{},
		},
	}
}

func (c *ComposeModel) AccessServiceName() string {
	return c.service.GetName()
}

func NewComposeModel(p *types.Project, l hclog.Logger) *ComposeModel {
	return &ComposeModel{
		composeProject: p,
		logger:         l,

		serviceaccount:        &corev1.ServiceAccount{},
		service:               &corev1.Service{},
		persistentvolumeclaim: &corev1.PersistentVolumeClaim{},
		deployment:            &appsv1.Deployment{},
	}
}

// objectsWithGVK packs all compose service dependant object into a single slice with all their GVKs
func (c *ComposeModel) objectsWithGVK() []map[schema.GroupVersionKind]client.Object {
	objects := []map[schema.GroupVersionKind]client.Object{
		{
			serviceAccountGVK: c.serviceaccount,
		},
		{
			serviceGVK: c.service,
		},
		{
			deploymentGVK: c.deployment,
		},
		{
			pvcGVK: c.persistentvolumeclaim,
		},
	}

	return objects
}

// fromCluster unpacks PluginRequest unstructured.Unstructured objects into client-go native structs
func (c *ComposeModel) fromCluster(objects []*unstructured.Unstructured) error {
	for _, obj := range objects {
		switch obj.GetKind() {
		case "ServiceAccount":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.serviceaccount); err != nil {
				return errors.Wrap(err, "error marshaling serviceaccount")
			}
		case "Service":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.service); err != nil {
				return errors.Wrap(err, "error marshaling service")
			}
		case "PersistentVolumeClaim":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.persistentvolumeclaim); err != nil {
				return errors.Wrap(err, "error marshaling persistentvolumeclaim")
			}
		case "Deployment":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.deployment); err != nil {
				return errors.Wrap(err, "error marshaling deployment")
			}
		default:
			return errUnknownObject
		}
	}
	return nil
}

func (c *ComposeModel) isReady() bool {
	if c.deployment == nil {
		return false
	}
	return c.deployment.Status.ReadyReplicas == c.deployment.Status.Replicas
}

// setObjects updates dependant object parameters according to PluginRequest
func (c *ComposeModel) setObjects(req *commons.PluginRequest) error {
	if err := c.updateServiceAccount(req); err != nil {
		return errors.Wrap(err, "error updating service account")
	}
	c.logger.Debug("serviceaccount updated", "object", c.serviceaccount)
	if err := c.updateServiceDeployment(req); err != nil {
		return errors.Wrap(err, "error updating service / deployment")
	}
	c.logger.Debug("persistentvolumeclaim updated", "object", c.persistentvolumeclaim)
	c.logger.Debug("deployment updated", "object", c.deployment)
	c.logger.Debug("service updated", "object", c.service)
	return nil
}

func (c *ComposeModel) updateServiceAccount(req *commons.PluginRequest) error {
	if c.serviceaccount.Name == "" {
		c.serviceaccount.Name = req.Name
		c.serviceaccount.Namespace = req.Namespace
	}
	c.serviceaccount.ObjectMeta.Labels = labels(req.Name)
	return nil
}

func (c *ComposeModel) updateServiceDeployment(req *commons.PluginRequest) error {
	if c.service.Name == "" {
		c.service.Name = req.Name
		c.service.Namespace = req.Namespace
	}
	c.service.Labels = labels(req.Name)
	c.service.Spec.Selector = labels(req.Name)
	c.service.Spec.Type = corev1.ServiceTypeClusterIP
	c.service.Spec.Ports = make([]corev1.ServicePort, 0)

	if c.deployment.Name == "" {
		c.deployment.Name = req.Name
		c.deployment.Namespace = req.Namespace
	}
	c.deployment.Labels = labels(req.Name)
	c.deployment.Spec.Replicas = &req.Replicas
	c.deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels(req.Name),
	}
	c.deployment.Spec.Template.Labels = labels(req.Name)
	c.deployment.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyAlways
	c.deployment.Spec.Template.Spec.Volumes = make([]corev1.Volume, 0)
	c.deployment.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        "127.0.0.1",
			Hostnames: []string{},
		},
	}
	c.deployment.Spec.Paused = false
	terminationGracePeriod := int64(60)
	c.deployment.Spec.Template.Spec.TerminationGracePeriodSeconds = &terminationGracePeriod

	containers := make([]corev1.Container, 0)
	// handle docker-compose services as deployment containers
	for _, composeService := range c.composeProject.Services {
		var container *corev1.Container
		for _, deploymentContainer := range c.deployment.Spec.Template.Spec.Containers {
			if deploymentContainer.Name == composeService.Name {
				container = &deploymentContainer
				c.logger.Debug("Deployment container found.", "object", container)
				break
			}
		}
		// append if not empty
		if container == nil {
			container = &corev1.Container{
				Name: composeService.Name,
			}
			c.logger.Debug("Deployment container not found. Creating one.", "object", container)
		}
		c.deployment.Spec.Template.Spec.HostAliases[0].Hostnames = append(c.deployment.Spec.Template.Spec.HostAliases[0].Hostnames, container.Name)

		imageValue, err := requestTemplatedValue(req, composeService.Image)
		if err != nil || imageValue == "" {
			return errors.Wrapf(err, "invalid image value: %s", imageValue)
		}
		container.Image = imageValue
		container.Command = composeService.Command

		container.Env = make([]corev1.EnvVar, 0)
		for env, val := range composeService.Environment {
			e := corev1.EnvVar{
				Name:  env,
				Value: "",
			}
			if val != nil {
				value, err := requestTemplatedValue(req, *val)
				if err != nil {
					return errors.Wrapf(err, "invalid env `%s` value: %s", e.Name, value)
				}
				e.Value = value
			}
			container.Env = append(container.Env, e)
		}
		sort.Slice(container.Env, func(i, j int) bool {
			return container.Env[i].Name < container.Env[j].Name
		})

		container.Ports = make([]corev1.ContainerPort, 0)
		for _, p := range composeService.Ports {
			target := intstr.FromInt(int(p.Target))
			proto := corev1.ProtocolTCP

			port := corev1.ContainerPort{
				Name:          target.String() + "-port",
				ContainerPort: target.IntVal,
				Protocol:      proto,
			}
			container.Ports = append(container.Ports, port)

			c.service.Spec.Ports = append(c.service.Spec.Ports, corev1.ServicePort{
				Name:       target.String() + "-port",
				Protocol:   proto,
				Port:       port.ContainerPort,
				TargetPort: target,
			})
		}
		sort.SliceStable(container.Ports, func(i, j int) bool {
			return container.Ports[i].Name < container.Ports[j].Name
		})

		container.VolumeMounts = make([]corev1.VolumeMount, 0)
		if len(composeService.Volumes) > 0 {
			if c.persistentvolumeclaim.GetName() == "" {
				c.persistentvolumeclaim.Name = req.Name
				c.persistentvolumeclaim.Namespace = req.Namespace
			}
			c.persistentvolumeclaim.Labels = labels(req.Namespace)

			c.persistentvolumeclaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			}
			c.persistentvolumeclaim.Spec.Resources = corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: resource.MustParse(req.VolumeSize),
				},
			}

			c.deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
				{
					Name: c.persistentvolumeclaim.GetName(),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: c.persistentvolumeclaim.GetName(),
							ReadOnly:  false,
						},
					},
				},
			}
		}
		for _, v := range composeService.Volumes {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      c.persistentvolumeclaim.GetName(),
				ReadOnly:  false,
				MountPath: v.Target,
				SubPath:   v.Source + "-" + container.Name,
			})
		}
		sort.SliceStable(container.VolumeMounts, func(i, j int) bool {
			return container.VolumeMounts[i].Name < container.VolumeMounts[j].Name
		})

		containers = append(containers, *container)
		c.logger.Debug("Deployment containers list", "containers", containers)
	}
	c.deployment.Spec.Template.Spec.Containers = containers

	sort.SliceStable(c.deployment.Spec.Template.Spec.Containers, func(i, j int) bool {
		return c.deployment.Spec.Template.Spec.Containers[i].Name < c.deployment.Spec.Template.Spec.Containers[j].Name
	})
	sort.SliceStable(c.deployment.Spec.Template.Spec.HostAliases[0].Hostnames, func(i, j int) bool {
		return c.deployment.Spec.Template.Spec.HostAliases[0].Hostnames[i] < c.deployment.Spec.Template.Spec.HostAliases[0].Hostnames[j]
	})
	return nil
}

func labels(name string) map[string]string {
	return map[string]string{
		"docker-compose.service/name": name,
	}
}

func requestTemplatedValue(req *commons.PluginRequest, value string) (string, error) {
	tmpl, err := template.New("value").Parse(value)
	if err != nil {
		return "", errors.Wrap(err, "error parsing template")
	}

	renderValues := &commons.PluginRequest{
		Name:       req.Name,
		Namespace:  req.Namespace,
		Host:       req.Host,
		Replicas:   req.Replicas,
		VolumeSize: req.VolumeSize,
		Version:    req.Version,
		TLSEnabled: req.TLSEnabled,
		Limits:     req.Limits,
		Parameters: req.Parameters,
		Objects:    nil,
	}
	data := &bytes.Buffer{}
	if err := tmpl.Execute(data, renderValues); err != nil {
		return "", errors.Wrap(err, "error rendering value")
	}
	return data.String(), nil
}
