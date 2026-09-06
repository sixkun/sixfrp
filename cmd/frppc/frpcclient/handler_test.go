package frpcclient

import (
	"os"
	"testing"

	frameworkconsole "github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/console"
	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
	"github.com/goravel/framework/log"

	"haokun-panel/cmd/frppc/config"
	"haokun-panel/utils"
)

// TestMain boots a minimal goravel foundation (config + log only) so
// facades.Log() is available; the handler logs skipped proxies through it.
// The full bootstrap can't be used here: its frppc provider imports this
// package, which would create an import cycle in the test binary.
func TestMain(m *testing.M) {
	foundation.Setup().
		WithCommands(func() []console.Command { return nil }).
		WithProviders(func() []contractsfoundation.ServiceProvider {
			return []contractsfoundation.ServiceProvider{
				&frameworkconsole.ServiceProvider{},
				&log.ServiceProvider{},
			}
		}).
		WithConfig(config.Boot).
		Create()
	os.Exit(m.Run())
}

// invalidHTTPProxy is an http proxy with neither subDomain nor customDomains,
// which fails frp's client-side validation. A valid tcp proxy sits alongside it.
const mixedValidityConfig = `
serverAddr = "127.0.0.1"
serverPort = 7000

[[proxies]]
name = "good-tcp"
type = "tcp"
localPort = 22
remotePort = 6000

[[proxies]]
name = "dev-http"
type = "http"
localPort = 8080
`

// TestNewClientHandlerSkipsInvalidProxy verifies that one invalid proxy does
// not take down the whole client: the valid proxy must survive.
func TestNewClientHandlerSkipsInvalidProxy(t *testing.T) {
	commonCfg, proxyCfgs, visitorCfgs, err := utils.LoadClientConfig([]byte(mixedValidityConfig), true)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(proxyCfgs) != 2 {
		t.Fatalf("expected 2 proxies loaded, got %d", len(proxyCfgs))
	}

	h, err := NewClientHandler(commonCfg, proxyCfgs, visitorCfgs)
	if err != nil {
		t.Fatalf("NewClientHandler should not fail on one invalid proxy: %v", err)
	}

	impl, ok := h.(*clientImpl)
	if !ok {
		t.Fatalf("unexpected handler type %T", h)
	}
	if len(impl.proxyCfgs) != 1 {
		t.Fatalf("expected 1 valid proxy to survive, got %d: %+v", len(impl.proxyCfgs), impl.proxyCfgs)
	}
	if _, ok := impl.proxyCfgs["good-tcp"]; !ok {
		t.Fatalf("valid proxy good-tcp was dropped: %+v", impl.proxyCfgs)
	}
	if _, ok := impl.proxyCfgs["dev-http"]; ok {
		t.Fatalf("invalid proxy dev-http should have been skipped")
	}
}
