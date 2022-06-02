/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"fmt"
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

func (g *PluginClient) callValidate(method string, args interface{}) *PluginResponseValidation {
	resp := &PluginResponseValidation{}
	err := g.client.Call("Plugin."+method, args, resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		//fmt.Println("Get()", err)
		panic(err)
	}

	return resp
}

func (g *PluginClient) Status(req PluginRequest) *PluginResponseStatus {
	resp := &PluginResponseStatus{}
	err := g.client.Call("Plugin.Status", req, resp)
	if err != nil {
		// You usually want your interfaces to return errors. If they don't,
		// there isn't much other choice here.
		//fmt.Println("Get()", err)
		panic(err)
	}

	return resp
}

func (g *PluginClient) Default() *PluginResponseDefault {
	fmt.Println("PluginClient -- Default -- begin")
	resp := &PluginResponseDefault{}
	err := g.client.Call("Plugin.Default", struct{}{}, resp)
	if err != nil {
		panic(err)
	}
	fmt.Println("PluginClient -- Default -- end")
	return resp
}

func (g *PluginClient) Convert(req PluginRequest) *PluginResponse {
	//fmt.Println("PluginClient -- Convert -- begin")
	res := g.call("Convert", req)
	//fmt.Println("PluginClient -- Convert -- end")
	return res
}

func (g *PluginClient) Types() *PluginResponse {
	return g.call("Types", struct{}{})
}

func (g *PluginClient) ValidateCreate(req PluginRequest) *PluginResponseValidation {
	return g.callValidate("ValidateCreate", req)
}

func (g *PluginClient) ValidateUpdate(req PluginRequest) *PluginResponseValidation {
	return g.callValidate("ValidateUpdate", req)
}

func (g *PluginClient) ValidateDelete(req PluginRequest) *PluginResponseValidation {
	return g.callValidate("ValidateDelete", req)
}

func (g *PluginClient) SetLogger(_ hclog.Logger) {
	panic("cannot executed")
}
