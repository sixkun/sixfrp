package frpcclient

import (
	"sync"
	"time"

	"github.com/fatedier/frp/client"
	"github.com/fatedier/frp/client/proxy"
	v1 "github.com/fatedier/frp/pkg/config/v1"
)

const (
	defaultRPCURL = "grpc://127.0.0.1:9001"
)

type options struct {
	enabled           bool
	clientID          string
	clientSecret      string
	rpcURL            string
	syncInterval      time.Duration
	enableRemoteShell bool
	totpSecret        string
}

// ClientHandler is the per-(clientID, serverID) frpc instance the controller
// manages. Implementations wrap a fatedier/frp client.Service.
type ClientHandler interface {
	Run()
	Stop()
	Running() bool
	Update([]v1.ProxyConfigurer, []v1.VisitorConfigurer)
	GetCommonCfg() *v1.ClientCommonConfig
	GetProxyStatus(name string) (*proxy.WorkingStatus, bool)
}

// clientImpl wraps a fatedier/frp client.Service.
type clientImpl struct {
	cli         *client.Service
	common      *v1.ClientCommonConfig
	proxyCfgs   map[string]v1.ProxyConfigurer
	visitorCfgs map[string]v1.VisitorConfigurer
	runningMu   sync.RWMutex
	running     bool
	done        chan struct{}
}
