/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package commons

import "github.com/hashicorp/go-plugin"

// HandshakeConfig are used to just do a basic handshake between
// a plugin and host. If the handshake fails, a user friendly error is shown.
// This prevents users from executing bad plugins or executing a plugin
// directory. It is a UX feature, not a security feature.
var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "KUBERLOGIC_SERVICE_PLUGIN",
	MagicCookieValue: "com.kuberlogic.service.plugin",
}
