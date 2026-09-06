package bootstrap

import (
	frppcproviders "haokun-panel/cmd/frppc/providers"

	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/event"
	"github.com/goravel/framework/log"
)

func Providers() []foundation.ServiceProvider {
	return []foundation.ServiceProvider{
		&console.ServiceProvider{},
		&log.ServiceProvider{},
		&event.ServiceProvider{},
		&frppcproviders.FrppcServiceProvider{},
	}
}
