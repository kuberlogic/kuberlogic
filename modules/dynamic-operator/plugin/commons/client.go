/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package commons

import (
	"github.com/hashicorp/go-hclog"
	"net/rpc"
)

var _ PluginService = &PluginClient{}

// Here is an implementation that talks over RPC
type PluginClient struct {
	client *rpc.Client
}

func (g *PluginClient) call(method string, args interface{}) *PluginResponse {
	resp := &PluginResponse{}
	err := g.client.Call("Plugin."+method, args, resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		//fmt.Println("Get()", err)
		panic(err)
	}

	return resp
}

func (g *PluginClient) Default() *PluginResponseDefault {
	resp := &PluginResponseDefault{}
	err := g.client.Call("Plugin.Default", struct{}{}, resp)
	if err != nil {
		panic(err)
	}

	return resp
}

func (g *PluginClient) Empty(req PluginRequest) *PluginResponse {
	return g.call("Empty", req)
}

func (g *PluginClient) ForUpdate(req PluginRequest) *PluginResponse {
	return g.call("ForUpdate", req)
}

func (g *PluginClient) ForCreate(req PluginRequest) *PluginResponse {
	return g.call("ForCreate", req)
}

func (g *PluginClient) Status(req PluginRequest) *PluginResponse {
	return g.call("Status", req)
}

func (g *PluginClient) Type() *PluginResponse {
	return g.call("Type", struct{}{})
}

func (g *PluginClient) ValidateCreate(req PluginRequest) *PluginResponse {
	return g.call("ValidateCreate", struct{}{})
}

func (g *PluginClient) ValidateUpdate(req PluginRequest) *PluginResponse {
	return g.call("ValidateUpdate", struct{}{})
}

func (g *PluginClient) ValidateDelete(req PluginRequest) *PluginResponse {
	return g.call("ValidateDelete", struct{}{})
}

func (g *PluginClient) SetLogger(_ hclog.Logger) {
	panic("cannot executed")
}
