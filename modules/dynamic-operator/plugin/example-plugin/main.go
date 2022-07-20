/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"encoding/gob"
	"github.com/hashicorp/go-hclog"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/example-plugin/plugin"
)

func main() {
	logger := hclog.New(&hclog.LoggerOptions{})
	svc := plugin.NewPostgresqlService(logger)
	commons.ServePlugin("postgresql", svc)
}

func init() {
	gob.Register(commons.PluginRequest{})
}
