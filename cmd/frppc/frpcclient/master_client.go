package frpcclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	gorillawebsocket "github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	appwsgrpc "cnb.cool/sixkun/sixfrp/v2/haokun/grpc/support/wsgrpc"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

func newMasterClient(rawRPCURL string) (pb.ClientMasterClient, *grpc.ClientConn, error) {
	rpcURL := strings.TrimSpace(rawRPCURL)
	if rpcURL == "" {
		rpcURL = defaultRPCURL
	}
	if !strings.Contains(rpcURL, "://") {
		if inferredWSURL, ok := inferWebSocketRPCURL(rpcURL); ok {
			rpcURL = inferredWSURL
		} else {
			rpcURL = "grpc://" + rpcURL
		}
	}

	u, err := url.Parse(rpcURL)
	if err != nil {
		return nil, nil, fmt.Errorf("parse rpc-url failed: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	host := u.Host
	if host == "" {
		host = u.Path
	}
	if host == "" {
		return nil, nil, fmt.Errorf("invalid rpc-url host: %s", rawRPCURL)
	}

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	target := host

	switch scheme {
	case "grpc", "":
	case "ws", "wss":
		wsURL := u
		if wsURL.Path == "" || wsURL.Path == "/" {
			wsURL.Path = "/grpc-ws"
		}
		target = "passthrough:///grpc-ws"
		dialOpts = append(dialOpts, grpc.WithContextDialer(websocketDialer(wsURL.String(), nil, true)))
	default:
		return nil, nil, fmt.Errorf("unsupported rpc-url scheme: %s", scheme)
	}

	conn, err := grpc.NewClient(target, dialOpts...)
	if err != nil {
		return nil, nil, fmt.Errorf("create grpc client failed: %w", err)
	}
	return pb.NewClientMasterClient(conn), conn, nil
}

func inferWebSocketRPCURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false
	}
	if strings.Contains(raw, "/grpc-ws") {
		return "ws://" + raw, true
	}

	host := raw
	if slash := strings.IndexByte(host, '/'); slash >= 0 {
		host = host[:slash]
	}

	if _, port, err := net.SplitHostPort(host); err == nil && port == "3000" {
		if strings.Contains(raw, "/") {
			return "ws://" + raw, true
		}
		return "ws://" + raw + "/grpc-ws", true
	}

	return "", false
}

func websocketDialer(wsURL string, header http.Header, insecureSkipVerify bool) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, _ string) (net.Conn, error) {
		dialer := gorillawebsocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify}}
		ws, _, err := dialer.DialContext(ctx, wsURL, header)
		if err != nil {
			return nil, err
		}
		return appwsgrpc.NewGorillaConn(ws), nil
	}
}
