package bootstrap

import (
	"haokun-panel/cmd/frppc/config"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
