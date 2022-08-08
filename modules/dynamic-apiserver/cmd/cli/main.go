/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/cli"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", filepath.Join(homedir.HomeDir(), ".kube", "config"))
	if err != nil {
		fmt.Println("Error building Kubernetes client: ", err)
		os.Exit(1)
	}

	k8sclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Println("Erro building Kubernetes client: ", err)
		os.Exit(1)
	}

	rootCmd, err := cli.MakeRootCmd(nil, k8sclient) // use default http client
	if err != nil {
		fmt.Println("Cmd construction error: ", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
