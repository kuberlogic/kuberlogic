// Code generated by go-swagger; DO NOT EDIT.

package main

import (
	"fmt"
	"os"

	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/generated/cli"
)

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

func main() {
	rootCmd, err := cli.MakeRootCmd()
	if err != nil {
		fmt.Println("Cmd construction error: ", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
