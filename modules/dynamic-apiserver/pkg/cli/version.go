/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import (
	dockerparser "github.com/novln/docker-parser"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func makeVersionCmd(k8sclient kubernetes.Interface) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Version of kuberlogic components",
		RunE:  runVersion(k8sclient),
	}
	return cmd
}

func runVersion(k8sclient kubernetes.Interface) func(command *cobra.Command, args []string) error {
	return func(command *cobra.Command, args []string) error {
		ctx := command.Context()
		list, err := k8sclient.CoreV1().Pods("kuberlogic").List(ctx, v1.ListOptions{
			LabelSelector: "control-plane=controller-manager",
		})
		if err != nil {
			return err
		}

		command.Printf("cli: %s\n", ver)
		for _, pod := range list.Items {
			for _, container := range pod.Spec.Containers {
				ref, err := dockerparser.Parse(container.Image)
				if err != nil {
					return err
				}
				command.Printf("%s: %s\n", container.Name, ref.Tag())
			}
			break
		}
		return nil
	}
}
