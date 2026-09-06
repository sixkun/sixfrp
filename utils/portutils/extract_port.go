package portutils

import (
	"encoding/json/v2"
	"strconv"
	"strings"

	v1 "github.com/fatedier/frp/pkg/config/v1"
)

// ExtractRemotePortFromTypedConfig returns the FRP remotePort value. Vhost
// proxies such as HTTP/HTTPS do not define one and therefore return zero.
func ExtractRemotePortFromTypedConfig(cfg v1.TypedProxyConfig) int {
	content, err := cfg.MarshalJSON()
	if err != nil {
		return 0
	}
	return ExtractRemotePortFromProxyConfig(cfg.GetBaseConfig().Type, nil, content)
}

// ExtractRemotePortFromProxyConfig extracts the remote_port from FRP proxy configuration
// Returns 0 if no remote port is found (e.g., for visitor proxies like stcp/sudp/xtcp)
func ExtractRemotePortFromProxyConfig(proxyType string, metas map[string]string, content []byte) int {
	// Visitor proxies don't use remote ports on server side
	switch strings.ToLower(proxyType) {
	case "stcp", "sudp", "xtcp":
		return 0
	}

	// Check metas for remotePort
	if remotePort := metas["remotePort"]; remotePort != "" {
		if port, err := strconv.Atoi(remotePort); err == nil && port > 0 {
			return port
		}
	}

	// Check metas for remote_port (snake_case)
	if remotePort := metas["remote_port"]; remotePort != "" {
		if port, err := strconv.Atoi(remotePort); err == nil && port > 0 {
			return port
		}
	}

	// Check JSON content
	if len(content) > 0 {
		var config map[string]interface{}
		if err := json.Unmarshal(content, &config); err == nil {
			// Try remote_port
			if remotePort, ok := config["remote_port"].(float64); ok && remotePort > 0 {
				return int(remotePort)
			}
			// Try remotePort (camelCase)
			if remotePort, ok := config["remotePort"].(float64); ok && remotePort > 0 {
				return int(remotePort)
			}
		}
	}

	return 0
}

// GetProtocolFromProxyType returns the network protocol (tcp/udp) for a proxy type
func GetProtocolFromProxyType(proxyType string) string {
	switch strings.ToLower(proxyType) {
	case "udp", "sudp":
		return "udp"
	case "tcp", "stcp", "xtcp", "http", "https":
		return "tcp"
	default:
		return "tcp" // default to tcp
	}
}
