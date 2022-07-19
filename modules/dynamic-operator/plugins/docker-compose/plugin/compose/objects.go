package compose

import (
	"fmt"
	"github.com/compose-spec/compose-go/types"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sort"
	"strconv"
)

const (
	ingressPathExtension = "x-kuberlogic-access-http-path"
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
		Kind:    "Ingress",
	}
	secretGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Secret",
	}
)

var (
	errUnknownObject          = errors.New("unknown object kind")
	errTooManyAccessPorts     = errors.New("too many access ports")
	errParsingPublishedPort   = errors.New("can't parse published port")
	errDuplicatePublishedPort = errors.New("duplicate published port")
	errIngressPathEmpty       = errors.New("HTTP access path is not found")
	errDuplicateIngressPath   = errors.New("HTTP access path has been already used")
)

type ComposeModel struct {
	composeProject *types.Project
	logger         hclog.Logger

	serviceaccount        *corev1.ServiceAccount
	service               *corev1.Service
	persistentvolumeclaim *corev1.PersistentVolumeClaim
	deployment            *appsv1.Deployment
	ingress               *networkingv1.Ingress
	secret                *corev1.Secret
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
		{
			ingressGVK: &networkingv1.Ingress{},
		},
		{
			secretGVK: &corev1.Secret{},
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
		ingress:               &networkingv1.Ingress{},
		secret:                &corev1.Secret{},
	}
}

// objectsWithGVK packs all compose service dependant object into a single slice with all their GVKs
func (c *ComposeModel) objectsWithGVK() []map[schema.GroupVersionKind]client.Object {
	return []map[schema.GroupVersionKind]client.Object{
		{
			serviceAccountGVK: c.serviceaccount,
		},
		{
			serviceGVK: c.service,
		},
		{
			secretGVK: c.secret,
		},
		{
			deploymentGVK: c.deployment,
		},
		{
			pvcGVK: c.persistentvolumeclaim,
		},
		{
			ingressGVK: c.ingress,
		},
	}
}

// fromCluster unpacks PluginRequest unstructured.Unstructured objects into client-go native structs
func (c *ComposeModel) fromCluster(objects []*unstructured.Unstructured) error {
	for _, obj := range objects {
		var object client.Object
		switch obj.GetKind() {
		case "ServiceAccount":
			object = c.serviceaccount
		case "Service":
			object = c.service
		case "PersistentVolumeClaim":
			object = c.persistentvolumeclaim
		case "Deployment":
			object = c.deployment
		case "Ingress":
			object = c.ingress
		case "Secret":
			object = c.secret
		default:
			return errUnknownObject
		}
		if err := commons.FromUnstructured(obj.UnstructuredContent(), object); err != nil {
			return errors.Wrapf(err, "error marshaling %s", obj.GetKind())
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
	if err := c.setApplicationObjects(req); err != nil {
		return errors.Wrap(err, "failed to set application objects")
	}
	c.logger.Debug("set persistentvolumeclaim", "object", c.persistentvolumeclaim)
	c.logger.Debug("set deployment", "object", c.deployment)
	c.logger.Debug("set serviceaccount", "object", c.serviceaccount)
	if err := c.setApplicationAccessObjects(req); err != nil {
		return errors.Wrap(err, "failed to set application access objects")
	}
	c.logger.Debug("set service", "object", c.service)
	c.logger.Debug("set ingress", "object", c.ingress)
	return nil
}

func (c *ComposeModel) setApplicationObjects(req *commons.PluginRequest) error {
	c.serviceaccount.Name = req.Name
	c.serviceaccount.Namespace = req.Namespace
	c.serviceaccount.ObjectMeta.Labels = labels(req.Name)

	c.deployment.Name = req.Name
	c.deployment.Namespace = req.Namespace

	c.deployment.Labels = labels(req.Name)
	c.deployment.Spec.Replicas = &req.Replicas
	c.deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels(req.Name),
	}
	c.deployment.Spec.Template.Labels = labels(req.Name)
	c.deployment.Spec.Template.Spec.ServiceAccountName = c.serviceaccount.GetName()
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

		vd := newViewData(req)
		// this will not be kept in secret even when a flag is set
		imageValue, _, err := vd.parse(composeService.Image)
		if err != nil || imageValue == "" {
			return errors.Wrapf(err, "invalid image value: %s", imageValue)
		}
		container.Image = imageValue
		container.Command = composeService.Command

		container.Env = make([]corev1.EnvVar, 0)
		for key, rawValue := range composeService.Environment {
			e := corev1.EnvVar{
				Name: key,
			}
			if rawValue != nil {
				value, keyId, err := vd.parse(*rawValue)
				if err != nil {
					return errors.Wrapf(err, "invalid key `%s` value: %s", e.Name, value)
				}

				if vd.isSecret(*rawValue) {
					c.secret.Name = req.Name

					// default secretId is composed of service name and env key
					secretKey := composeService.Name + "_" + key
					// custom keyId is set, use this instead of existing one
					if keyId != "" {
						secretKey = keyId
					}

					// create a new key when it is not set
					if _, ok := c.secret.Data[secretKey]; !ok {
						if c.secret.StringData == nil {
							c.secret.StringData = make(map[string]string)
						}
						c.secret.StringData[secretKey] = value
					}

					e.ValueFrom = &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: c.secret.GetName(),
							},
							Key: secretKey,
						},
					}
				} else {
					e.Value = value
				}
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
		}
		sort.SliceStable(container.Ports, func(i, j int) bool {
			return container.Ports[i].Name < container.Ports[j].Name
		})

		container.VolumeMounts = make([]corev1.VolumeMount, 0)
		if len(composeService.Volumes) > 0 {
			c.persistentvolumeclaim.Name = req.Name
			c.persistentvolumeclaim.Namespace = req.Namespace
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

func (c *ComposeModel) setApplicationAccessObjects(req *commons.PluginRequest) error {
	c.service.Name = req.Name
	c.service.Namespace = req.Namespace

	c.service.Labels = labels(req.Name)
	c.service.Spec.Selector = labels(req.Name)
	c.service.Spec.Type = corev1.ServiceTypeClusterIP
	c.service.Spec.Ports = []corev1.ServicePort{}

	svcPorts := make(map[int32]string, 0)

	for _, svc := range c.composeProject.Services {
		if svc.Ports == nil || len(svc.Ports) == 0 {
			continue
		}
		if len(svc.Ports) > 1 {
			return errors.Wrap(errTooManyAccessPorts, fmt.Sprintf("service %s must have only one published port", svc.Name))
		}

		published := svc.Ports[0]
		targetPort := intstr.FromInt(int(published.Target))
		publishedPort, err := strconv.Atoi(published.Published)
		if err != nil {
			return errors.Wrap(errParsingPublishedPort, fmt.Sprintf("can't parse port %s", published.Published))
		}
		if _, found := svcPorts[int32(publishedPort)]; found {
			return errors.Wrap(errDuplicatePublishedPort, fmt.Sprintf("port %d is already exposed", publishedPort))
		}

		svcPort := corev1.ServicePort{
			Name:       "app-" + targetPort.String(),
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(publishedPort),
			TargetPort: targetPort,
		}
		c.service.Spec.Ports = append(c.service.Spec.Ports, svcPort)

		path := "/"
		if svc.Extensions != nil && svc.Extensions[ingressPathExtension] != nil {
			path = svc.Extensions[ingressPathExtension].(string)
		}

		svcPorts[svcPort.Port] = path
	}
	sort.Slice(c.service.Spec.Ports, func(i, j int) bool {
		return c.service.Spec.Ports[i].Name < c.service.Spec.Ports[j].Name
	})

	// now handle ingress
	// Host is not specified, no ingress object
	if req.Host == "" {
		return nil
	}

	c.ingress.Name = req.Name
	c.ingress.Namespace = req.Namespace
	c.ingress.Labels = labels(req.Name)

	// add TLS if specified
	if req.TLSEnabled {
		c.ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{req.Host},
				SecretName: req.TLSSecretName,
			},
		}
	}

	ingressPaths := make(map[string]interface{}, 0)
	paths := make([]networkingv1.HTTPIngressPath, 0)
	pathType := networkingv1.PathTypePrefix
	for _, port := range c.service.Spec.Ports {
		path, found := svcPorts[port.Port]
		if !found {
			return errors.Wrap(errIngressPathEmpty, fmt.Sprintf("ingress path not found for port %s", port.Name))
		}

		if _, found := ingressPaths[path]; found {
			return errors.Wrap(errDuplicateIngressPath, fmt.Sprintf("duplicate ingress path `%s`", path))
		}

		ingressPath := networkingv1.HTTPIngressPath{
			PathType: &pathType,
			Path:     path,
			Backend: networkingv1.IngressBackend{
				Service: &networkingv1.IngressServiceBackend{
					Name: c.service.GetName(),
					Port: networkingv1.ServiceBackendPort{
						Name: port.Name,
					},
				},
			},
		}
		ingressPaths[path] = nil
		paths = append(paths, ingressPath)
	}

	c.ingress.Spec.Rules = []networkingv1.IngressRule{
		{
			Host: req.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		},
	}
	return nil
}

func labels(name string) map[string]string {
	return map[string]string{
		"docker-compose.service/name": name,
	}
}
