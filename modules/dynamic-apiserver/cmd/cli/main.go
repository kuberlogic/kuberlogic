/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"fmt"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/pkg/cli"
	"os"
)

func main() {
	rootCmd, err := cli.MakeRootCmd(nil) // make default http client
	if err != nil {
		fmt.Println("Cmd construction error: ", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
