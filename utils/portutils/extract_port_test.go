package portutils

import (
	"testing"

	v1 "github.com/fatedier/frp/pkg/config/v1"
	"github.com/stretchr/testify/require"
)

func TestExtractRemotePortFromTypedConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "tcp mirrors remote port",
			content: `{"type":"tcp","name":"tcp-test","localIP":"127.0.0.1","localPort":80,"remotePort":9000}`,
			want:    9000,
		},
		{
			name:    "http has no remote port",
			content: `{"type":"http","name":"http-test","localIP":"127.0.0.1","localPort":80,"customDomains":["example.com"]}`,
			want:    0,
		},
		{
			name:    "https has no remote port",
			content: `{"type":"https","name":"https-test","localIP":"127.0.0.1","localPort":443,"customDomains":["example.com"]}`,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg v1.TypedProxyConfig
			require.NoError(t, cfg.UnmarshalJSON([]byte(tt.content)))
			require.Equal(t, tt.want, ExtractRemotePortFromTypedConfig(cfg))
		})
	}
}
