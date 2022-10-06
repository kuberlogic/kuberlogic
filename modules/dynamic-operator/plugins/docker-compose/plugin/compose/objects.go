package compose

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/compose-spec/compose-go/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
)

const (
	IngressPathExtension       = "x-kuberlogic-access-http-path"
	HealthEndpointExtension    = "x-kuberlogic-health-endpoint"
	SetCredentialsCmdExtension = "x-kuberlogic-set-credentials-cmd"
	ConfigsExtension           = "x-kuberlogic-file-configs"
	SecretsExtension           = "x-kuberlogic-secrets"
)

var (
	serviceAccountGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ServiceAccount",
	}
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
	configmapGVK = schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "ConfigMap",
	}
)

var (
	ErrUnknownObject                = errors.New("unknown object kind")
	ErrTooManyAccessPorts           = errors.New("only one exposed port is allowed")
	ErrParsingPublishedPort         = errors.New("can't parse published port")
	ErrDuplicatePublishedPort       = errors.New("duplicate published port")
	ErrIngressPathEmpty             = errors.New("HTTP access path is not found")
	ErrDuplicateIngressPath         = errors.New("HTTP access path has been already used")
	ErrTooManyCredentialsCommands   = errors.New("too many " + SetCredentialsCmdExtension + " extensions")
	ErrCredentialsCommandNotDefined = errors.New(SetCredentialsCmdExtension + " extension not found")
	ErrConfigsDecodeFailed          = errors.New(ConfigsExtension + " must be of type map[string]string")
	ErrSecretsDecodeFailed          = errors.New(SecretsExtension + " must be of type map[string]string")
	ErrStringConversionFailed       = errors.New("failed to read string")
)

type ComposeModel struct {
	composeProject *types.Project
	logger         *zap.SugaredLogger

	serviceaccount        *corev1.ServiceAccount
	service               *corev1.Service
	persistentvolumeclaim *corev1.PersistentVolumeClaim
	deployment            *appsv1.Deployment
	ingress               *networkingv1.Ingress
	secret                *corev1.Secret
	configmap             *corev1.ConfigMap
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
		{
			configmapGVK: &corev1.ConfigMap{},
		},
	}
}

func (c *ComposeModel) AccessServiceName() string {
	return c.service.GetName()
}

func NewComposeModel(p *types.Project, l *zap.SugaredLogger) *ComposeModel {
	return &ComposeModel{
		composeProject: p,
		logger:         l,

		serviceaccount:        &corev1.ServiceAccount{},
		service:               &corev1.Service{},
		persistentvolumeclaim: &corev1.PersistentVolumeClaim{},
		deployment:            &appsv1.Deployment{},
		ingress:               &networkingv1.Ingress{},
		secret:                &corev1.Secret{},
		configmap:             &corev1.ConfigMap{},
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
		{
			configmapGVK: c.configmap,
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
		case "ConfigMap":
			object = c.configmap
		default:
			return ErrUnknownObject
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
	c.serviceaccount.SetName(req.Name)
	c.serviceaccount.SetNamespace(req.Namespace)
	c.serviceaccount.ObjectMeta.SetLabels(labels(req.Name))

	c.secret.SetName(req.Name)
	c.secret.SetNamespace(req.Namespace)
	c.secret.SetLabels(labels(req.Name))
	if c.secret.Data == nil {
		c.secret.Data = make(map[string][]byte, 0)
	}
	// now go and set secret data
	if secrets, set := c.composeProject.Extensions[SecretsExtension]; set {
		for k, v := range secrets.(map[string]interface{}) {
			if _, set := c.secret.Data[k]; set {
				c.logger.Debug("secret %s is already set. skipping.", k)
				continue
			}

			value, err := req.RenderTemplate(v.(string), c.secret.Data)
			if err != nil {
				return errors.Wrapf(err, "failed to generate secret %s", k)
			}
			c.secret.Data[k] = []byte(value.String())
		}
	}

	c.deployment.SetName(req.Name)
	c.deployment.SetNamespace(req.Namespace)
	c.deployment.SetLabels(labels(req.Name))

	c.deployment.Spec.Strategy.Type = appsv1.RecreateDeploymentStrategyType
	c.deployment.Spec.Replicas = &req.Replicas
	c.deployment.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: labels(req.Name),
	}
	c.deployment.Spec.Template.SetLabels(labels(req.Name))
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

		// this will not be kept in secret even when a flag is set
		imageValue, err := req.RenderTemplate(composeService.Image, c.secret.Data)
		if err != nil || imageValue.String() == "" {
			return errors.Wrapf(err, "invalid image value: %s", imageValue.String())
		}
		container.Image = imageValue.String()
		container.Command = composeService.Command

		if container.Env, err = c.buildContainerEnvVars(&composeService, req); err != nil {
			return errors.Wrapf(err, "failed to build environment variables for service %s", composeService.Name)
		}

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

			healthz := "/"
			if customHealthz := composeService.Extensions[HealthEndpointExtension]; customHealthz != nil {
				healthz = customHealthz.(string)
			}

			container.ReadinessProbe = &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: healthz,
						Port: intstr.FromString(port.Name),
					},
				},
				FailureThreshold:    3,
				InitialDelaySeconds: 5,
				PeriodSeconds:       5,
			}
		}
		sort.SliceStable(container.Ports, func(i, j int) bool {
			return container.Ports[i].Name < container.Ports[j].Name
		})

		if container.VolumeMounts, err = c.buildContainerVolumeMounts(&composeService, req); err != nil {
			return errors.Wrapf(err, "failed to build volume mounts for container %s", composeService.Name)
		}

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
	c.service.SetName(req.Name)
	c.service.SetNamespace(req.Namespace)
	c.service.SetLabels(labels(req.Name))

	c.service.Spec.Selector = labels(req.Name)
	c.service.Spec.Type = corev1.ServiceTypeClusterIP
	c.service.Spec.Ports = []corev1.ServicePort{}

	svcPorts := make(map[int32]string, 0)

	for _, svc := range c.composeProject.Services {
		if svc.Ports == nil || len(svc.Ports) == 0 {
			continue
		}

		published := svc.Ports[0]
		targetPort := intstr.FromInt(int(published.Target))
		publishedPort, err := strconv.Atoi(published.Published)
		if err != nil {
			return errors.Wrap(ErrParsingPublishedPort, fmt.Sprintf("can't render port %s", published.Published))
		}

		svcPort := corev1.ServicePort{
			Name:       "app-" + targetPort.String(),
			Protocol:   corev1.ProtocolTCP,
			Port:       int32(publishedPort),
			TargetPort: targetPort,
		}
		c.service.Spec.Ports = append(c.service.Spec.Ports, svcPort)

		path := "/"
		if svc.Extensions != nil && svc.Extensions[IngressPathExtension] != nil {
			path = svc.Extensions[IngressPathExtension].(string)
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

	c.ingress.SetName(req.Name)
	c.ingress.SetNamespace(req.Namespace)
	c.ingress.SetLabels(labels(req.Name))

	if req.IngressClass != "" {
		c.ingress.Spec.IngressClassName = &req.IngressClass
	}

	// add TLS if specified
	if !req.Insecure {
		c.ingress.Spec.TLS = []networkingv1.IngressTLS{
			{
				Hosts:      []string{req.Host},
				SecretName: req.TLSSecretName,
			},
		}
	}

	paths := make([]networkingv1.HTTPIngressPath, 0)
	pathType := networkingv1.PathTypePrefix
	for _, port := range c.service.Spec.Ports {
		path, found := svcPorts[port.Port]
		if !found {
			return errors.Wrap(ErrIngressPathEmpty, fmt.Sprintf("ingress path not found for port %s", port.Name))
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

func (c *ComposeModel) GetCredentialsMethod(req *commons.PluginRequestCredentialsMethod) (*commons.PluginResponseCredentialsMethod, error) {
	// search across services
	var commandTemplate, container string
	for _, svc := range c.composeProject.Services {
		if cmdTmpl := svc.Extensions[SetCredentialsCmdExtension]; cmdTmpl != nil {
			commandTemplate, container = cmdTmpl.(string), svc.Name
			break
		}
	}

	if container == "" || commandTemplate == "" {
		return nil, ErrCredentialsCommandNotDefined
	}

	v, err := req.RenderTemplate(commandTemplate)
	if err != nil {
		return nil, err
	}

	return &commons.PluginResponseCredentialsMethod{
		Method: "exec",
		Exec: commons.CredentialsMethodExec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: labels(req.Name),
			},
			Container: container,
			Command:   strings.Split(v.String(), " "),
		},
	}, nil
}

func labels(name string) map[string]string {
	return map[string]string{
		"docker-compose.service/name": name,
	}
}

// buildContainerEnvVars transforms composeSvc environment variables to Kuberlogic compatible []corev1.EnvVar
func (c *ComposeModel) buildContainerEnvVars(composeSvc *types.ServiceConfig, req *commons.PluginRequest) ([]corev1.EnvVar, error) {
	envs := make([]corev1.EnvVar, 0)

	for key, rawValue := range composeSvc.Environment {
		e := corev1.EnvVar{
			Name:  key,
			Value: "",
		}
		if rawValue != nil {
			value, err := req.RenderTemplate(*rawValue, c.secret.Data)
			if err != nil {
				return nil, errors.Wrapf(err, "invalid key `%s` value: %s", e.Name, value.String())
			}

			if value.SecretID != "" {
				// use secretKeyRef instead of raw value
				e.ValueFrom = &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: c.secret.GetName(),
						},
						Key: value.SecretID,
					},
				}
			} else {
				e.Value = value.String()
			}
		}

		envs = append(envs, e)
	}
	// all additional parameters mapped to env vars for each container
	c.logger.Debugf("extra parameters: %+v", req.Parameters)
	for key, value := range req.Parameters {
		envs = append(envs, corev1.EnvVar{
			Name:  key,
			Value: fmt.Sprintf("%v", value),
		})
	}

	sort.Slice(envs, func(i, j int) bool {
		return envs[i].Name < envs[j].Name
	})
	return envs, nil
}

func (c *ComposeModel) buildContainerVolumeMounts(s *types.ServiceConfig, req *commons.PluginRequest) ([]corev1.VolumeMount, error) {
	volumeMounts := make([]corev1.VolumeMount, 0)

	if len(s.Volumes) > 0 {
		c.persistentvolumeclaim.SetName(req.Name)
		c.persistentvolumeclaim.SetNamespace(req.Namespace)
		c.persistentvolumeclaim.Labels = labels(req.Namespace)

		c.persistentvolumeclaim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
			corev1.ReadWriteOnce,
		}
		if req.StorageClass != "" {
			c.persistentvolumeclaim.Spec.StorageClassName = &req.StorageClass
		}

		limits, err := req.GetLimits()
		if err != nil {
			return nil, errors.Wrap(err, "failed to get limits")
		}

		storage := limits.Storage()
		if storage != nil && !storage.IsZero() {
			c.persistentvolumeclaim.Spec.Resources = corev1.ResourceRequirements{
				Requests: map[corev1.ResourceName]resource.Quantity{
					corev1.ResourceStorage: *storage,
				},
			}
		}

		var found bool
		for _, v := range c.deployment.Spec.Template.Spec.Volumes {
			if v.Name == c.persistentvolumeclaim.GetName() {
				found = true
			}
		}
		if !found {
			c.deployment.Spec.Template.Spec.Volumes = append(c.deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: c.persistentvolumeclaim.GetName(),
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: c.persistentvolumeclaim.GetName(),
							ReadOnly:  false,
						},
					},
				})
		}
	}

	if configMap, set := s.Extensions[ConfigsExtension]; set {
		c.configmap.SetName(req.Name)
		c.configmap.SetNamespace(req.Namespace)
		c.configmap.Data = make(map[string]string, 0)

		if c.configmap.Annotations == nil {
			c.configmap.Annotations = make(map[string]string, 0)
		}

		const configVolumeName = "file-configs"
		var found bool

		for _, v := range c.deployment.Spec.Template.Spec.Volumes {
			if v.Name == configVolumeName {
				found = true
			}
		}

		if !found {
			c.deployment.Spec.Template.Spec.Volumes = append(c.deployment.Spec.Template.Spec.Volumes,
				corev1.Volume{
					Name: configVolumeName,
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: c.configmap.GetName(),
							},
						},
					},
				})
		}

		for path, config := range configMap.(map[string]interface{}) {
			rendered, err := req.RenderTemplate(config.(string), c.secret.Data)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to render config %s", path)
			}

			md5Path := md5.Sum([]byte(path))
			normalizedPath := hex.EncodeToString(md5Path[:])

			c.configmap.Annotations[normalizedPath] = path
			c.configmap.Data[normalizedPath] = rendered.String()

			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      configVolumeName,
				MountPath: path,
				SubPath:   normalizedPath,
			})
		}
	}

	for _, v := range s.Volumes {
		volumeMounts = append(volumeMounts, corev1.VolumeMount{
			Name:      c.persistentvolumeclaim.GetName(),
			ReadOnly:  false,
			MountPath: v.Target,
			SubPath:   v.Source + "-" + s.Name,
		})
	}
	sort.SliceStable(volumeMounts, func(i, j int) bool {
		return volumeMounts[i].Name < volumeMounts[j].Name
	})
	return volumeMounts, nil
}

// ValidateComposeProject reads docker-compose project p, checks it against Kuberlogic requirements and returns error when validation fails
func ValidateComposeProject(p *types.Project) error {
	// validate secrets extension
	if secrets, set := p.Extensions[SecretsExtension]; set {
		if _, converted := secrets.(map[string]interface{}); !converted {
			return errors.Wrapf(ErrSecretsDecodeFailed, "failed to decode parameter %s", SecretsExtension)
		}
		for k, v := range secrets.(map[string]interface{}) {
			if _, ok := v.(string); !ok {
				return errors.Wrapf(ErrSecretsDecodeFailed, "it is expected that key `%s` value is string", k)
			}
		}
	}

	// validate configs
	for _, svc := range p.Services {
		if configs, set := svc.Extensions[ConfigsExtension]; set {
			if _, converted := configs.(map[string]interface{}); !converted {
				return errors.Wrapf(ErrConfigsDecodeFailed, "failed to decode parameter %s in service %s", ConfigsExtension, svc.Name)
			}

			for k, v := range configs.(map[string]interface{}) {
				if _, ok := v.(string); !ok {
					return errors.Wrapf(ErrConfigsDecodeFailed, " it is expected that key `%s` value is string", k)
				}
			}
		}
	}

	// ingressPaths contains ingresses, we will check for duplicates
	// svcPorts contains exposed ports, we will check for duplicates
	ingressPaths := make(map[string]string, 0)
	svcPorts := make(map[string]string, 0)
	for _, svc := range p.Services {
		// no ports found
		if svc.Ports == nil || len(svc.Ports) == 0 {
			continue
		}

		// too many ports found
		if len(svc.Ports) != 1 {
			return errors.Wrapf(ErrTooManyAccessPorts, "error in service `%s`", svc.Name)
		}

		port := svc.Ports[0]
		if name, found := svcPorts[port.Published]; found {
			return errors.Wrapf(ErrDuplicatePublishedPort,
				"failed to expose service `%s` on port `%s`, it is already used by service `%s`", svc.Name, port.Published, name)
		}
		svcPorts[port.Published] = svc.Name

		// check for ingress paths duplicates
		path := "/"
		if customPath, found := svc.Extensions[IngressPathExtension]; found {
			var converted bool
			if path, converted = customPath.(string); !converted {
				return errors.Wrapf(ErrStringConversionFailed, "failed to decode parameter `%s` in service `%s`", IngressPathExtension, svc.Name)
			}
		}

		if name, found := ingressPaths[path]; found {
			return errors.Wrapf(ErrDuplicateIngressPath,
				"failed to expose service `%s` on HTTP path `%s`, it is already used by service `%s`", svc.Name, path, name)
		}
		ingressPaths[path] = svc.Name
	}

	// check for update credentials method
	var container, commandTemplate string
	for _, svc := range p.Services {
		if val := svc.Extensions[SetCredentialsCmdExtension]; val != nil {
			if container != "" || commandTemplate != "" {
				return errors.Wrapf(ErrTooManyCredentialsCommands, "only one service can have `%s` parameter set", SetCredentialsCmdExtension)
			}
			if _, ok := val.(string); !ok {
				return errors.Wrapf(ErrStringConversionFailed, "failed to decode parameter `%s` in service `%s`", SetCredentialsCmdExtension, svc.Name)
			}
			container, commandTemplate = svc.Name, val.(string)
		}
	}
	return nil
}
