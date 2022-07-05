package compose

import (
	"github.com/compose-spec/compose-go/types"
	"github.com/go-test/deep"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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
			Name:       "demo-kls",
			Namespace:  "demo-kls",
			Host:       "demo.example.com",
			Replicas:   1,
			VolumeSize: "1G",
			Version:    "",
		}

		c := NewComposeModel(testValidProject, hclog.L())

		It("Should create valid Kubernetes objects", func() {
			By("Checking Reconcile return parameters")
			objs, err := c.Reconcile(requests)
			Expect(err).Should(BeNil())
			Expect(len(objs)).Should(Equal(4))

			By("Validating returned Deployment")
			firstDeployment := *c.deployment
			podSpec := c.deployment.Spec.Template.Spec
			Expect(len(podSpec.Containers)).Should(Equal(2))
			container := podSpec.Containers[0]
			Expect(container.Name).Should(Equal("demo-app"))
			Expect(len(container.Env)).Should(Equal(4))
			Expect(container.Ports[0].ContainerPort).Should(Equal(int32(80)))
			Expect(podSpec.Containers[1].Name).Should(Equal("demo-db"))
			Expect(podSpec.Containers[0]).ShouldNot(Equal(podSpec.Containers[1]))

			By("Validating returned Service")
			serviceSpec := c.service.Spec
			Expect(len(serviceSpec.Ports)).Should(Equal(1))
			port := serviceSpec.Ports[0]
			Expect(port.Port).Should(Equal(int32(80)))

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
			Expect(len(secondRunObjs)).Should(Equal(4))

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

		c := NewComposeModel(testInvalidProject, hclog.L())

		requests := &commons.PluginRequest{
			Name:       "demo-kls",
			Namespace:  "demo-kls",
			Host:       "demo.example.com",
			Replicas:   1,
			VolumeSize: "1G",
			Version:    "",
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

		c := NewComposeModel(testInvalidProject, hclog.L())

		requests := &commons.PluginRequest{
			Name:       "demo-kls",
			Namespace:  "demo-kls",
			Host:       "demo.example.com",
			Replicas:   1,
			VolumeSize: "1G",
			Version:    "whatever",
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

			c := NewComposeModel(testProject, hclog.L())

			requests := &commons.PluginRequest{
				Name:       "demo-kls",
				Namespace:  "demo-kls",
				Host:       "demo.example.com",
				Replicas:   1,
				VolumeSize: "1G",
				Version:    "whatever",
			}

			By("Checking Reconcile return parameters")
			It("Should reconcile without errors", func() {
				_, err := c.Reconcile(requests)
				Expect(err).Should(BeNil())

				By("Checking persistentvolumeclaim object")
				Expect(c.persistentvolumeclaim).ShouldNot(BeNil())

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
	})
})
