/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import (
	"encoding/gob"
	"github.com/hashicorp/go-plugin"
	"net/rpc"
)

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
	gob.Register(&PluginResponse{})
	gob.Register(&PluginResponseDefault{})
	gob.Register(&PluginResponseValidation{})
	gob.Register(&PluginResponseStatus{})

	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}
