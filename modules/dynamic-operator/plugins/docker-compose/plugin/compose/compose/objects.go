package compose

import (
	"bytes"
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	v12 "k8s.io/api/apps/v1"
	v13 "k8s.io/api/core/v1"
	v14 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	ingressGVK = schema.GroupVersionKind{
		Group:   "networking.k8s.io",
		Version: "v1",
		Kind:    "Ingress"}
)

var (
	errUnknownObject   = errors.New("unknown object kind")
	errVolumeNotFound  = errors.New("requested volume definition not found")
	errTooManySvcPorts = errors.New("too many ports defined in service")
)

type ComposeModel struct {
	composeProject *types.Project
	logger         hclog.Logger

	serviceaccount         *v13.ServiceAccount
	service                *v13.Service
	persistentvolumeclaims map[string]*v13.PersistentVolumeClaim
	deployment             *v12.Deployment
	ingress                *v14.Ingress
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
			serviceGVK: &v13.Service{},
		},
		{
			serviceAccountGVK: &v13.ServiceAccount{},
		},
		{
			pvcGVK: &v13.PersistentVolumeClaim{},
		},
		{
			deploymentGVK: &v12.Deployment{},
		},
		{
			ingressGVK: &v14.Ingress{},
		},
	}
}

func NewComposeModel(p *types.Project, l hclog.Logger) *ComposeModel {
	return &ComposeModel{
		composeProject: p,
		logger:         l,

		serviceaccount:         &v13.ServiceAccount{},
		service:                &v13.Service{},
		persistentvolumeclaims: make(map[string]*v13.PersistentVolumeClaim, 0),
		deployment:             &v12.Deployment{},
		ingress:                &v14.Ingress{},
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
	}
	if c.ingress != nil {
		objects = append(objects, map[schema.GroupVersionKind]client.Object{ingressGVK: c.ingress})
	}

	for _, pvc := range c.persistentvolumeclaims {
		objects = append(objects, map[schema.GroupVersionKind]client.Object{
			pvcGVK: pvc,
		})
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
			pvc := &v13.PersistentVolumeClaim{}
			if err := commons.FromUnstructured(obj.UnstructuredContent(), pvc); err != nil {
				return errors.Wrap(err, "error marshaling persistentvolumeclaim")
			}
			c.persistentvolumeclaims[pvc.GetName()] = pvc
		case "Deployment":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.deployment); err != nil {
				return errors.Wrap(err, "error marshaling deployment")
			}
		case "Ingress":
			if err := commons.FromUnstructured(obj.UnstructuredContent(), c.ingress); err != nil {
				return errors.Wrap(err, "error marshaling ingress")
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
	if err := c.updatePVCs(req); err != nil {
		return errors.Wrap(err, "error updating persistentvolumeclaims")
	}
	c.logger.Debug("pvcs updated", "objects", c.persistentvolumeclaims)
	if err := c.updateServiceDeployment(req); err != nil {
		return errors.Wrap(err, "error updating service / deployment")
	}
	c.logger.Debug("deployment updated", "object", c.deployment)
	c.logger.Debug("service updated", "object", c.service)
	if err := c.updateIngress(req); err != nil {
		return errors.Wrap(err, "error updating ingress")
	}
	c.logger.Debug("ingress updated", "object", c.ingress)
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

func (c *ComposeModel) updatePVCs(req *commons.PluginRequest) error {
	for name := range c.composeProject.Volumes {
		vol, found := c.persistentvolumeclaims[name]
		if !found {
			vol = &v13.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{
					Name:      name,
					Namespace: req.Namespace,
				},
			}
		}
		vol.Labels = labels(req.Name)
		vol.Spec.AccessModes = []v13.PersistentVolumeAccessMode{
			v13.ReadWriteOnce,
		}
		vol.Spec.Resources = v13.ResourceRequirements{
			Requests: map[v13.ResourceName]resource.Quantity{
				v13.ResourceStorage: resource.MustParse(req.VolumeSize),
			},
		}
		c.persistentvolumeclaims[name] = vol
	}
	return nil
}

func (c *ComposeModel) updateServiceDeployment(req *commons.PluginRequest) error {
	if c.service.Name == "" {
		c.service.Name = req.Name
		c.service.Namespace = req.Namespace
	}
	c.service.Labels = labels(req.Name)
	c.service.Spec.Selector = labels(req.Name)
	c.service.Spec.Type = v13.ServiceTypeClusterIP
	c.service.Spec.Ports = make([]v13.ServicePort, 0)

	if c.deployment.Name == "" {
		c.deployment.Name = req.Name
		c.deployment.Namespace = req.Namespace
	}
	c.deployment.Labels = labels(req.Name)
	c.deployment.Spec.Replicas = &req.Replicas
	c.deployment.Spec.Selector = &v1.LabelSelector{
		MatchLabels: labels(req.Name),
	}
	c.deployment.Spec.Template.Labels = labels(req.Name)
	c.deployment.Spec.Template.Spec.RestartPolicy = v13.RestartPolicyAlways
	c.deployment.Spec.Template.Spec.Volumes = make([]v13.Volume, 0)
	c.deployment.Spec.Template.Spec.HostAliases = make([]v13.HostAlias, 0)
	c.deployment.Spec.Paused = false

	// handle deployment volumes
	for _, vol := range c.persistentvolumeclaims {
		c.deployment.Spec.Template.Spec.Volumes = append(c.deployment.Spec.Template.Spec.Volumes, v13.Volume{
			Name: vol.GetName(),
			VolumeSource: v13.VolumeSource{
				PersistentVolumeClaim: &v13.PersistentVolumeClaimVolumeSource{
					ClaimName: vol.GetName(),
					ReadOnly:  false,
				},
			},
		})
	}
	sort.Slice(c.deployment.Spec.Template.Spec.Volumes, func(i, j int) bool {
		return c.deployment.Spec.Template.Spec.Volumes[i].Name < c.deployment.Spec.Template.Spec.Volumes[j].Name
	})

	containers := make([]v13.Container, 0)
	// handle docker-compose services as deployment containers
	for _, composeService := range c.composeProject.Services {
		var container *v13.Container
		for _, deploymentContainer := range c.deployment.Spec.Template.Spec.Containers {
			if deploymentContainer.Name == composeService.Name {
				container = &deploymentContainer
				c.logger.Debug("Deployment container found.", "object", container)
				break
			}
		}
		// append if not empty
		if container == nil {
			container = &v13.Container{
				Name: composeService.Name,
			}
			c.logger.Debug("Deployment container not found. Creating one.", "object", container)
		}
		c.deployment.Spec.Template.Spec.HostAliases = append(c.deployment.Spec.Template.Spec.HostAliases, v13.HostAlias{
			IP:        "127.0.0.1",
			Hostnames: []string{container.Name},
		})

		imageValue, err := requestTemplatedValue(req, composeService.Image)
		if err != nil || imageValue == "" {
			return errors.Wrapf(err, "invalid image value: %s", imageValue)
		}
		container.Image = imageValue
		container.Command = composeService.Command

		container.Env = make([]v13.EnvVar, 0)
		for env, val := range composeService.Environment {
			e := v13.EnvVar{
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

		container.Ports = make([]v13.ContainerPort, 0)
		for _, p := range composeService.Ports {
			target := intstr.FromInt(int(p.Target))
			proto := v13.ProtocolTCP

			port := v13.ContainerPort{
				Name:          target.String() + "-port",
				ContainerPort: target.IntVal,
				Protocol:      proto,
			}
			container.Ports = append(container.Ports, port)

			c.service.Spec.Ports = append(c.service.Spec.Ports, v13.ServicePort{
				Name:       target.String() + "-port",
				Protocol:   proto,
				Port:       port.ContainerPort,
				TargetPort: target,
			})
		}

		container.VolumeMounts = make([]v13.VolumeMount, 0)
		for _, v := range composeService.Volumes {
			pvc, found := c.persistentvolumeclaims[v.Source]
			if !found {
				// pvc not found, error
				return errVolumeNotFound
			}
			container.VolumeMounts = append(container.VolumeMounts, v13.VolumeMount{
				Name:      pvc.Name,
				ReadOnly:  false,
				MountPath: v.Target,
				SubPath:   "data",
			})
		}

		containers = append(containers, *container)
		c.logger.Debug("Deployment containers list", "containers", containers)
	}
	c.deployment.Spec.Template.Spec.Containers = containers
	return nil
}

func (c *ComposeModel) updateIngress(req *commons.PluginRequest) error {
	// Host is not set
	if req.Host == "" {
		c.ingress = nil
		return nil
	}

	if len(c.service.Spec.Ports) != 1 {
		return errTooManySvcPorts
	}
	if c.ingress.Name == "" {
		c.ingress.Name = req.Name
		c.ingress.Namespace = req.Namespace
	}

	pathType := v14.PathTypePrefix
	c.ingress.Labels = labels(req.Name)
	c.ingress.Spec.Rules = []v14.IngressRule{
		{
			Host: req.Host,
			IngressRuleValue: v14.IngressRuleValue{HTTP: &v14.HTTPIngressRuleValue{
				Paths: []v14.HTTPIngressPath{
					{
						Path:     "/",
						PathType: &pathType,
						Backend: v14.IngressBackend{
							Service: &v14.IngressServiceBackend{
								Name: c.service.Name,
								Port: v14.ServiceBackendPort{
									Name: c.service.Spec.Ports[0].Name,
								},
							},
							Resource: nil,
						},
					},
				},
			}},
		},
	}
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
	data := &bytes.Buffer{}
	if err := tmpl.Execute(data, req); err != nil {
		return "", errors.Wrap(err, "error rendering value")
	}
	return data.String(), nil
}
