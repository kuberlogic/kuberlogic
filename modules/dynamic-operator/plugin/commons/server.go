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

//var _ PluginService = &PluginServer{}

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
