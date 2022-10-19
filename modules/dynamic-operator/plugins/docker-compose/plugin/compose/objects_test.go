package compose

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/types"
	"github.com/go-test/deep"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var envVal = "val"

var _ = Describe("docker-compose model", func() {
	Context("When reconciling demo project", func() {
		testValidProject := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Extensions: map[string]interface{}{
						"x-kuberlogic-health-endpoint": "/health",
					},
					Environment: types.MappingWithEquals{
						"DEMO_ENV": nil,
						"ENV1":     &envVal,
						"ENV2":     &envVal,
						"ENV3":     &envVal,
					},
					Image: "demo:test",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}

		requests := &commons.PluginRequest{
			Name:      "demo-kls",
			Namespace: "demo-kls",
			Replicas:  1,
			Version:   "",
		}

		c := NewComposeModel(testValidProject, zap.NewRaw().Sugar())

		It("Should create valid Kubernetes objects", func() {
			By("Checking Reconcile return parameters")
			objs, err := c.Reconcile(requests)
			Expect(err).Should(BeNil())
			Expect(len(objs)).Should(Equal(6))

			By("Validating returned Deployment")
			firstDeployment := *c.deployment
			podSpec := c.deployment.Spec.Template.Spec
			Expect(len(podSpec.Containers)).Should(Equal(2))
			container := podSpec.Containers[0]
			Expect(container.Name).Should(Equal("demo-app"))
			Expect(len(container.Env)).Should(Equal(4))
			Expect(container.Ports[0].ContainerPort).Should(Equal(int32(80)))
			Expect(container.ReadinessProbe).ShouldNot(BeNil())
			Expect(container.ReadinessProbe.HTTPGet.Path).Should(Equal("/health"))
			Expect(container.ReadinessProbe.HTTPGet.Port.String()).Should(Equal(container.Ports[0].Name))
			Expect(podSpec.Containers[1].Name).Should(Equal("demo-db"))
			Expect(podSpec.Containers[0]).ShouldNot(Equal(podSpec.Containers[1]))

			By("Validating returned Service")
			serviceSpec := c.service.Spec
			Expect(len(serviceSpec.Ports)).Should(Equal(1))
			port := serviceSpec.Ports[0]
			Expect(port.Port).Should(Equal(int32(8001)))

			// populate requests objects from the previous run
			for _, elem := range objs {
				for gvk, obj := range elem {
					unstructuredObj, _ := commons.ToUnstructured(obj, gvk)
					requests.AddObject(unstructuredObj)
				}
			}
			By("Checking Reconcile result for the 2nd time")
			secondRunObjs, secondErr := c.Reconcile(requests)
			Expect(secondErr).Should(BeNil())
			Expect(len(secondRunObjs)).Should(Equal(6))

			By("Validating returned Deployment")
			secondDeployment := *c.deployment
			Expect(c.deployment.Spec.Template.Spec.Containers[0].Name).Should(Equal("demo-app"))
			Expect(c.deployment.Spec.Template.Spec.Containers[1].Name).Should(Equal("demo-db"))
			Expect(len(c.deployment.Spec.Template.Spec.Containers)).Should(Equal(2))
			Expect(len(c.deployment.Spec.Template.Spec.Containers[0].Env)).Should(Equal(4))
			Expect(c.deployment.Spec.Template.Spec.Containers[0].Env[0].Name).Should(Equal("DEMO_ENV"))
			Expect(c.deployment.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).Should(Equal(int32(80)))

			deploymentsEqual := deep.Equal(firstDeployment.Spec, secondDeployment.Spec)
			Expect(deploymentsEqual).Should(BeNil())

			By("Ingress should not be set with empty Host")
			Expect(c.ingress.GetName()).Should(Equal(""))
		})
	})

	Context("When reconciling project with invalid template", func() {
		testInvalidProject := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Environment: types.MappingWithEquals{
						"DEMO_ENV": nil,
						"ENV1":     &envVal,
						"ENV2":     &envVal,
						"ENV3":     &envVal,
					},
					Image: "demo:{{ .test",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}

		c := NewComposeModel(testInvalidProject, zap.NewRaw().Sugar())

		requests := &commons.PluginRequest{
			Name:      "demo-kls",
			Namespace: "demo-kls",
			Host:      "demo.example.com",
			Replicas:  1,
			Version:   "",
		}

		It("Should reconcile without errors", func() {
			_, err := c.Reconcile(requests)
			Expect(err).ShouldNot(BeNil())
		})
	})

	Context("When reconciling project with valid template", func() {
		testInvalidProject := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Environment: types.MappingWithEquals{
						"DEMO_ENV": nil,
						"ENV1":     &envVal,
						"ENV2":     &envVal,
						"ENV3":     &envVal,
					},
					Image: "demo:{{ .Version }}",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}

		c := NewComposeModel(testInvalidProject, zap.NewRaw().Sugar())

		requests := &commons.PluginRequest{
			Name:      "demo-kls",
			Namespace: "demo-kls",
			Host:      "demo.example.com",
			Replicas:  1,
			Version:   "whatever",
		}

		It("Should reconcile without errors", func() {
			_, err := c.Reconcile(requests)
			Expect(err).Should(BeNil())

			By("Checking container image")
			Expect(c.deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal("demo:whatever"))
		})
	})

	Context("When Reconciling with two volumes", func() {
		Context("When reconciling project with valid template", func() {
			testProject := &types.Project{
				Name:       "test",
				WorkingDir: "/tmp",
				Services: types.Services{
					types.ServiceConfig{
						Name:    "demo-app",
						Command: types.ShellCommand{"cmd", "arg"},
						Environment: types.MappingWithEquals{
							"DEMO_ENV": nil,
							"ENV1":     &envVal,
							"ENV2":     &envVal,
							"ENV3":     &envVal,
						},
						Image: "demo:{{ .Version }}",
						Ports: []types.ServicePortConfig{
							{
								Target:    80,
								Published: "8001",
							},
						},
						Volumes: []types.ServiceVolumeConfig{
							{
								Source: "demo",
								Target: "/tmp/demo",
							},
							{
								Source: "second",
								Target: "/tmp/second",
							},
						},
					},
					types.ServiceConfig{
						Name:    "demo-db",
						Command: types.ShellCommand{"cmd", "arg"},
						Image:   "demodb:test",
					},
				},
				Networks: nil,
				Volumes: types.Volumes{
					"demo": types.VolumeConfig{
						Name: "demo",
					},
					"second": types.VolumeConfig{
						Name: "second",
					},
				},
			}

			c := NewComposeModel(testProject, zap.NewRaw().Sugar())

			requests := &commons.PluginRequest{
				Name:         "demo-kls",
				Namespace:    "demo-kls",
				Host:         "demo.example.com",
				Replicas:     1,
				Version:      "whatever",
				StorageClass: "default",
			}
			err := requests.SetLimits(&corev1.ResourceList{
				corev1.ResourceStorage: resource.MustParse("10G"),
			})
			Expect(err).Should(BeNil())

			By("Checking Reconcile return parameters")
			It("Should reconcile without errors", func() {
				_, err := c.Reconcile(requests)
				Expect(err).Should(BeNil())

				By("Checking persistentvolumeclaim object")
				Expect(c.persistentvolumeclaim).ShouldNot(BeNil())
				Expect(*c.persistentvolumeclaim.Spec.Resources.Requests.Storage()).Should(Equal(resource.MustParse("10G")))
				Expect(*c.persistentvolumeclaim.Spec.StorageClassName).Should(Equal(requests.StorageClass))

				By("Checking pod volume")
				Expect(len(c.deployment.Spec.Template.Spec.Volumes)).Should(Equal(1))

				By("Checking container volume mounts")
				Expect(len(c.deployment.Spec.Template.Spec.Containers[0].VolumeMounts)).Should(Equal(2))
				Expect(c.deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0]).Should(Equal(corev1.VolumeMount{
					Name:      c.persistentvolumeclaim.GetName(),
					ReadOnly:  false,
					MountPath: "/tmp/demo",
					SubPath:   "demo-demo-app",
				}))
				Expect(c.deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1]).Should(Equal(corev1.VolumeMount{
					Name:      c.persistentvolumeclaim.GetName(),
					ReadOnly:  false,
					MountPath: "/tmp/second",
					SubPath:   "second-demo-app",
				}))
			})
		})

		Context("When Reconciling with two volumes", func() {
			Context("When reconciling project with valid template", func() {
				testProject := &types.Project{
					Name:       "test",
					WorkingDir: "/tmp",
					Services: types.Services{
						types.ServiceConfig{
							Name:    "demo-app",
							Command: types.ShellCommand{"cmd", "arg"},
							Environment: types.MappingWithEquals{
								"DEMO_ENV": nil,
								"ENV1":     &envVal,
								"ENV2":     &envVal,
								"ENV3":     &envVal,
							},
							Image: "demo:{{ .Version }}",
							Ports: []types.ServicePortConfig{
								{
									Target:    80,
									Published: "8001",
								},
							},
							Volumes: []types.ServiceVolumeConfig{
								{
									Source: "demo",
									Target: "/tmp/demo",
								},
								{
									Source: "second",
									Target: "/tmp/second",
								},
							},
						},
						types.ServiceConfig{
							Name:    "demo-db",
							Command: types.ShellCommand{"cmd", "arg"},
							Image:   "demodb:test",
						},
					},
					Networks: nil,
					Volumes: types.Volumes{
						"demo": types.VolumeConfig{
							Name: "demo",
						},
						"second": types.VolumeConfig{
							Name: "second",
						},
					},
				}

				requests := &commons.PluginRequest{
					Name:      "demo-kls",
					Namespace: "demo-kls",
					Host:      "demo.example.com",
					Replicas:  1,
					Version:   "whatever",
				}

				It("Should create an ingress with a valid HTTP project", func() {
					mReq := *requests
					mReq.IngressClass = "default"
					mReq.TLSSecretName = ""
					mReq.Insecure = true

					m := *testProject
					m.Services[0].Ports = []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					}
					m.Services[0].Extensions = map[string]interface{}{
						"x-kuberlogic-access-http-path": "/api",
					}
					m.Services[1].Ports = []types.ServicePortConfig{
						{
							Target:    90,
							Published: "8002",
						},
					}

					c := NewComposeModel(&m, zap.NewRaw().Sugar())
					_, err := c.Reconcile(&mReq)
					Expect(err).Should(BeNil())
					By("Checking Ingress object")
					Expect(c.ingress.GetName()).Should(Equal(mReq.Name))
					Expect(*c.ingress.Spec.IngressClassName).Should(Equal(mReq.IngressClass))
					Expect(len(c.ingress.Spec.TLS)).Should(Equal(0))
					Expect(len(c.ingress.Spec.Rules)).Should(Equal(1))
					Expect(c.ingress.Spec.Rules[0].Host).Should(Equal(mReq.Host))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP).ShouldNot(BeNil())
					Expect(len(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths)).Should(Equal(2))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Path).Should(Equal("/api"))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Name).Should(Equal(c.service.GetName()))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[0].Backend.Service.Port.Name).Should(Equal(c.service.Spec.Ports[0].Name))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Path).Should(Equal("/"))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Backend.Service.Name).Should(Equal(c.service.GetName()))
					Expect(c.ingress.Spec.Rules[0].IngressRuleValue.HTTP.Paths[1].Backend.Service.Port.Name).Should(Equal(c.service.Spec.Ports[1].Name))
				})
				It("Should create an ingress with a valid HTTPS project", func() {
					mReq := *requests
					mReq.TLSSecretName = "demo"
					mReq.Insecure = false

					m := *testProject
					m.Services[0].Ports = []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					}
					m.Services[0].Extensions = map[string]interface{}{
						"x-kuberlogic-access-http-path": "/api",
					}
					m.Services[1].Ports = []types.ServicePortConfig{
						{
							Target:    90,
							Published: "8002",
						},
					}

					c := NewComposeModel(&m, zap.NewRaw().Sugar())
					_, err := c.Reconcile(&mReq)
					Expect(err).Should(BeNil())
					By("Checking Ingress object")
					Expect(c.ingress.Spec.TLS[0].SecretName).Should(Equal(mReq.TLSSecretName))
				})
			})
		})
	})
	Context("When reconciling with templates env, keep secrets, cache", func() {
		env1 := `{{ "abc" }}`
		env2 := `{{ GenerateKey 30 }}`
		env3 := `{{ Secret "key" }}`
		env4 := `{{ Secret "key" }}`
		testValidProject := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Extensions: map[string]interface{}{
				"x-kuberlogic-secrets": map[string]interface{}{
					"token": "exampletoken",
					"key":   "{{ GenerateRSA 2048 | Base64 }}",
				},
			},
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Environment: types.MappingWithEquals{
						"ENV1": &env1,
						"ENV2": &env2,
						"ENV3": &env3,
						"ENV4": &env4,
					},
					Image: "demo:test",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
					Environment: map[string]*string{
						"ENV1": &env1,
						"ENV4": &env4,
					},
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}

		requests := &commons.PluginRequest{
			Name:      "demo-kls",
			Namespace: "demo-kls",
			Replicas:  1,
			Version:   "",
		}

		c := NewComposeModel(testValidProject, zap.NewRaw().Sugar())

		It("Should create valid Kubernetes objects", func() {
			By("Checking Reconcile return parameters")
			objs, err := c.Reconcile(requests)
			Expect(err).Should(BeNil())
			Expect(len(objs)).Should(Equal(6))

			By("Validating returned Deployment")

			podSpec := c.deployment.Spec.Template.Spec
			Expect(len(podSpec.Containers)).Should(Equal(2))
			container := podSpec.Containers[0]
			Expect(len(container.Env)).Should(Equal(4))

			Expect(container.Env[0].Value).Should(Equal("abc"))
			genKey := container.Env[1].Value
			Expect(len(genKey)).Should(Equal(30))
			Expect(container.Env[2].ValueFrom.SecretKeyRef.Key).Should(Equal("key"))
			Expect(*container.Env[2].ValueFrom).Should(Equal(*container.Env[3].ValueFrom))

			By("Validating returned secret")
			Expect(c.secret.Name).Should(Equal("demo-kls"))
			Expect(len(c.secret.Data)).Should(Equal(2))
			genRSA := c.secret.Data["key"]
			Expect(len(string(genRSA))).Should(BeNumerically(">", 2048))

			// validate secrets
			Expect(c.secret.Data["token"]).Should(Equal([]byte("exampletoken")))

			By("Repopulating with existing objects")
			// populate requests objects from the previous run
			for _, elem := range objs {
				for gvk, obj := range elem {
					unstructuredObj, _ := commons.ToUnstructured(obj, gvk)
					requests.AddObject(unstructuredObj)
				}
			}

			By("Checking Reconcile result for the 2nd time")
			_, secondErr := c.Reconcile(requests)
			Expect(secondErr).Should(BeNil())

			By("Validating returned Deployment")
			cont := c.deployment.Spec.Template.Spec.Containers[0]
			Expect(len(cont.Env)).Should(Equal(4))
			Expect(container.Env[2].ValueFrom.SecretKeyRef.Key).Should(Equal("key"))
			Expect(*container.Env[2].ValueFrom).Should(Equal(*container.Env[3].ValueFrom))

			By("Validating returned secret")
			Expect(c.secret.Name).Should(Equal("demo-kls"))
			Expect(c.secret.Data["key"]).Should(Equal(genRSA))
		})
	})
	Context("When GetCredentialsMethod is called", func() {
		proj := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demo:test",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}
		requests := &commons.PluginRequestCredentialsMethod{
			Name: "demo-kls",
		}

		It("Should fail when extension is not set", func() {
			c := NewComposeModel(proj, zap.NewRaw().Sugar())
			resp, err := c.GetCredentialsMethod(requests)
			Expect(resp).Should(BeNil())
			Expect(err).Should(Equal(ErrCredentialsCommandNotDefined))
		})

		It("Should render correct credentials command", func() {
			requests.Data = map[string]string{
				"user":     "admin",
				"password": "demopwd",
			}

			proj.Services[1].Extensions = map[string]interface{}{}
			proj.Services[0].Extensions = map[string]interface{}{"x-kuberlogic-set-credentials-cmd": "cli set password --user {{ .user }} --password {{ .password }}"}
			expectedContainer := proj.Services[0].Name
			expectedCommand := "cli set password --user admin --password demopwd"

			c := NewComposeModel(proj, zap.NewRaw().Sugar())
			resp, err := c.GetCredentialsMethod(requests)
			Expect(err).Should(BeNil())
			Expect(resp).ShouldNot(BeNil())
			Expect(resp.Method).Should(Equal("exec"))
			Expect(resp.Exec.Command).Should(Equal(strings.Split(expectedCommand, " ")))
			Expect(resp.Exec.Container).Should(Equal(expectedContainer))
		})
	})
	Context("When configs extension is set", func() {
		proj := &types.Project{
			Name:       "test",
			WorkingDir: "/tmp",
			Services: types.Services{
				types.ServiceConfig{
					Name:    "demo-app",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demo:test",
					Ports: []types.ServicePortConfig{
						{
							Target:    80,
							Published: "8001",
						},
					},
					Volumes: []types.ServiceVolumeConfig{
						{
							Source: "demo",
							Target: "/tmp/demo",
						},
					},
				},
				types.ServiceConfig{
					Name:    "demo-db",
					Command: types.ShellCommand{"cmd", "arg"},
					Image:   "demodb:test",
				},
			},
			Networks: nil,
			Volumes: types.Volumes{
				"demo": types.VolumeConfig{
					Name: "demo",
				},
			},
		}
		requests := &commons.PluginRequest{
			Name: "demo-kls",
		}

		It("Should be mapped to container correctly", func() {
			proj.Services[0].Extensions = map[string]interface{}{
				"x-kuberlogic-file-configs": map[string]interface{}{
					"/test": "app={{ .Name }}",
				},
			}
			c := NewComposeModel(proj, zap.NewRaw().Sugar())
			resp, err := c.Reconcile(requests)
			Expect(resp).ShouldNot(BeNil())
			Expect(err).Should(BeNil())

			md5Key := md5.Sum([]byte("/test"))
			expectedKey := hex.EncodeToString(md5Key[:])

			Expect(len(c.configmap.Data)).Should(Equal(1))
			Expect(c.configmap.Data[expectedKey]).Should(Equal("app=demo-kls"))
			Expect(c.deployment.Spec.Template.Spec.Volumes[1].ConfigMap.Name).Should(Equal(c.configmap.GetName()))
			Expect(c.deployment.Spec.Template.Spec.Volumes[1].Name).Should(Equal("file-configs"))
			Expect(c.deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name).Should(Equal(c.deployment.Spec.Template.Spec.Volumes[1].Name))
			Expect(c.deployment.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath).Should(Equal("/test"))
		})
		It("Should be using a secret to container correctly", func() {
			proj.Extensions = map[string]interface{}{
				"x-kuberlogic-secrets": map[string]interface{}{
					"app-key": "{{ GenerateKey 30 }}",
				},
			}
			proj.Services[0].Extensions = map[string]interface{}{
				"x-kuberlogic-file-configs": map[string]interface{}{
					"/test": `app={{ Secret "app-key" }}`,
				},
			}
			c := NewComposeModel(proj, zap.NewRaw().Sugar())
			resp, err := c.Reconcile(requests)
			Expect(resp).ShouldNot(BeNil())
			Expect(err).Should(BeNil())

			md5Key := md5.Sum([]byte("/test"))
			expectedKey := hex.EncodeToString(md5Key[:])

			Expect(len(c.configmap.Data)).Should(Equal(1))
			Expect(c.configmap.Data[expectedKey]).Should(MatchRegexp("app=[a-zA-Z]+"))
		})
	})

	Context("validating docker-compose project", func() {
		When("it is valid", func() {
			It("Should succeed", func() {
				p := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
						},
					},
				}

				Expect(ValidateComposeProject(p)).Should(BeNil())
			})
		})

		When("it is not valid", func() {
			It("should fail with incorrect secrets configuration", func() {
				q := &types.Project{
					Name: "test",
					Extensions: map[string]interface{}{
						"x-kuberlogic-secrets": "test",
					},
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrSecretsDecodeFailed)).Should(BeTrue())
			})

			It("should fail with incorrect configs", func() {
				q := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
							Extensions: map[string]interface{}{
								"x-kuberlogic-file-configs": "demo",
							},
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrConfigsDecodeFailed))
			})

			It("should fail when two ports are published", func() {
				q := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
							Ports: []types.ServicePortConfig{
								{
									Published: "1",
									Target:    1,
								},
								{
									Published: "2",
									Target:    2,
								},
							},
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrTooManyAccessPorts)).Should(BeTrue())
			})

			It("should fail when two services expose the same port", func() {
				port := 1

				q := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
							Ports: []types.ServicePortConfig{
								{
									Published: strconv.Itoa(port),
									Target:    uint32(port),
								},
							},
						},
						{
							Name:  "demo1",
							Image: "demo1",
							Ports: []types.ServicePortConfig{
								{
									Published: strconv.Itoa(port),
									Target:    uint32(port),
								},
							},
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrDuplicatePublishedPort)).Should(BeTrue())
			})

			It("should fail when two services share the same ingress path", func() {
				q := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
							Ports: []types.ServicePortConfig{
								{
									Published: "1",
									Target:    1,
								},
							},
						},
						{
							Name:  "demo1",
							Image: "demo1",
							Ports: []types.ServicePortConfig{
								{
									Published: "2",
									Target:    2,
								},
							},
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrDuplicateIngressPath)).Should(BeTrue())
			})

			It("should fail when two services have update-credentials-defined", func() {
				q := &types.Project{
					Name: "test",
					Services: []types.ServiceConfig{
						{
							Name:  "demo",
							Image: "demo",
							Ports: []types.ServicePortConfig{
								{
									Published: "1",
									Target:    1,
								},
							},
							Extensions: map[string]interface{}{
								"x-kuberlogic-set-credentials-cmd": "test",
							},
						},
						{
							Name:  "demo1",
							Image: "demo1",
							Extensions: map[string]interface{}{
								"x-kuberlogic-set-credentials-cmd": "test",
							},
						},
					},
				}

				Expect(errors.Is(ValidateComposeProject(q), ErrTooManyCredentialsCommands)).Should(BeTrue())
			})
		})
	})
})
