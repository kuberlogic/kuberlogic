package commons

import (
	"encoding/gob"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// PluginService is the interface that we're exposing as a plugin.
type PluginService interface {
	Empty(req PluginRequest) *PluginResponse
	ForCreate(req PluginRequest) *PluginResponse
	ForUpdate(req PluginRequest) *PluginResponse
	Status(req PluginRequest) *PluginResponse
	Type() *PluginResponse

	Default() *PluginResponseDefault
	ValidateCreate(req PluginRequest) *PluginResponse
	ValidateUpdate(req PluginRequest) *PluginResponse
	ValidateDelete(req PluginRequest) *PluginResponse
}

type PluginRequestEmpty struct{}
type PluginRequest struct {
	Name      string
	Namespace string

	Replicas   int32
	VolumeSize string
	Version    string

	Parameters map[string]interface{}
	Object     *unstructured.Unstructured
}

type PluginResponse struct {
	Object  *unstructured.Unstructured
	IsReady bool
	Error   error
}

type PluginResponseDefault struct {
	Replicas   int32
	VolumeSize string
	Version    string
	Parameters map[string]interface{}

	Error error
}

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

// Here is the RPC server that PluginClient talks to, conforming to
// the requirements of net/rpc
type PluginServer struct {
	// This is the real implementation
	Impl PluginService
}

func (s *PluginServer) Empty(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.Empty(req)
	return nil
}

func (s *PluginServer) ForCreate(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.ForCreate(req)
	return nil
}

func (s *PluginServer) ForUpdate(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.ForUpdate(req)
	return nil
}

func (s *PluginServer) Status(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.Status(req)
	return nil
}

func (s *PluginServer) Type(_ PluginRequestEmpty, resp *PluginResponse) error {
	*resp = *s.Impl.Type()
	return nil
}

func (s *PluginServer) Default(_ PluginRequestEmpty, resp *PluginResponseDefault) error {
	*resp = *s.Impl.Default()
	return nil
}

func (s *PluginServer) ValidateCreate(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}

func (s *PluginServer) ValidateUpdate(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}

func (s *PluginServer) ValidateDelete(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a PluginServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return PluginClient for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type Plugin struct {
	// Impl Injection
	Impl PluginService
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginServer{Impl: p.Impl}, nil
}

func (Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginClient{client: c}, nil
}

func init() {
	//gob.Register(error)
	gob.Register(&PluginResponse{})
	gob.Register(&PluginResponseDefault{})
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}
