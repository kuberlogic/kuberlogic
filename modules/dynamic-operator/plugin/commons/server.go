/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

//var _ PluginService = &PluginServer{}

// Here is the RPC server that PluginClient talks to, conforming to
// the requirements of net/rpc
type PluginServer struct {
	// This is the real implementation
	Impl PluginService
}

func (s *PluginServer) Convert(req PluginRequest, resp *PluginResponse) error {
	*resp = *s.Impl.Convert(req)
	return nil
}

func (s *PluginServer) Status(req PluginRequest, resp *PluginResponseStatus) error {
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

func (s *PluginServer) ValidateCreate(req PluginRequest, resp *PluginResponseValidation) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}

func (s *PluginServer) ValidateUpdate(req PluginRequest, resp *PluginResponseValidation) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}

func (s *PluginServer) ValidateDelete(req PluginRequest, resp *PluginResponseValidation) error {
	*resp = *s.Impl.ValidateCreate(req)
	return nil
}
