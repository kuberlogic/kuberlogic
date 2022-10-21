/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	"fmt"
	"net/http"
	"strings"

	client2 "github.com/go-openapi/runtime/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
)

func makeInfoCmd(k8sclient kubernetes.Interface, apiClientFunc func() (*client.ServiceAPI, error)) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Info about kuberlogic components",
		RunE:  runInfo(k8sclient, apiClientFunc),
	}
	return cmd
}

func runInfo(k8sclient kubernetes.Interface, apiClientFunc func() (*client.ServiceAPI, error)) func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		ctx := command.Context()

		isApiserverWorking := true
		globalStatus := true

		apiClient, err := apiClientFunc()
		if err != nil {
			return err
		}
		params := service.NewServiceListParams()
		response, err := apiClient.Service.ServiceList(params,
			client2.APIKeyAuth("X-Token", "header", viper.GetString(tokenFlag)))
		if err != nil {
			isApiserverWorking = false
			globalStatus = false
		}

		klPods, err := k8sclient.CoreV1().Pods("kuberlogic").List(ctx, v1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil {
			return err
		}

		for _, pod := range klPods.Items {
			for _, status := range pod.Status.ContainerStatuses {
				containerStatus := "Ready"
				if !status.Ready {
					globalStatus = false
					containerStatus = "NOT Ready"
				}
				command.Printf("%s: %s\n", status.Name, containerStatus)
			}
		}

		command.Println()
		hostname := viper.GetString(apiHostFlag)
		scheme := viper.GetString(schemeFlag)

		if isApiserverWorking {
			_ = response.GetPayload()
			command.Printf("API server is running at %s://%s\n", scheme, hostname)
		} else {
			command.Printf("API server is NOT running at %s://%s\n", scheme, hostname)
		}

		chargebeeEndpoint, err := getKuberlogicServiceEndpoint("kls-chargebee-integration", 1, k8sclient)
		if err != nil {
			command.Printf("Failed to get ChargeBee integration service endpoint")
			return err
		}
		fullChargebeeEndpoint := fmt.Sprintf("http://%s", chargebeeEndpoint)
		if _, err := http.Get(fullChargebeeEndpoint); err == nil {
			command.Printf("ChargeBee integration service is running at %s\n", fullChargebeeEndpoint)
		} else {
			command.Printf("ChargeBee integration service is NOT running at %s\n", fullChargebeeEndpoint)
		}

		var klSecretConfigName string
		for _, pod := range klPods.Items {
			for _, c := range pod.Spec.Containers {
				if c.Name == "chargebee-integration" {
					for _, e := range c.Env {
						if e.Name == strings.ToUpper(installChargebeeIntegrationUser) {
							klSecretConfigName = e.ValueFrom.SecretKeyRef.Name
						}
					}
				}
			}
		}
		klSecretConfig, err := k8sclient.CoreV1().Secrets("kuberlogic").Get(ctx, klSecretConfigName, v1.GetOptions{})
		if err != nil {
			command.Println("Error retrieving Kuberlogic config: " + err.Error())
		}
		command.Println("ChargeBee webhook user: ", klSecretConfig.Data[strings.ToUpper(installChargebeeIntegrationUser)])
		command.Println("ChargeBee webhook password: ", klSecretConfig.Data[strings.ToUpper(installChargebeeIntegrationPassword)])

		command.Println()
		if globalStatus {
			command.Println("Status: Ready")
		} else {
			command.Println("Status: NOT Ready")
		}

		command.Println()
		command.Printf("To further debug and diagnose problems, use '%s diag'.\n", exeName)
		return nil
	}
}
