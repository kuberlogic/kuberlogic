/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	client2 "github.com/go-openapi/runtime/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/generated/client/service"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
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

		list, err := k8sclient.CoreV1().Pods("kuberlogic").List(ctx, v1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil {
			return err
		}

		for _, pod := range list.Items {
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

		command.Println()
		if globalStatus {
			command.Println("Status: Ready")
		} else {
			command.Println("Status: NOT Ready")
		}

		command.Println()
		command.Println("To further debug and diagnose problems, use './kuberlogic diag'.")
		return nil
	}
}
