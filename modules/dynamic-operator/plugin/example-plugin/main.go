/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package main

import (
	"encoding/gob"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/commons"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/plugin/example-plugin/plugin"
)

func main() {
	commons.ServePlugin("postgresql", &plugin.PostgresqlService{})
}

func init() {
	gob.Register(commons.PluginRequest{})
}
