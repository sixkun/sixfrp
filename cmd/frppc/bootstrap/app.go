package bootstrap

import (
	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/config"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation"
)

func Boot() contractsfoundation.Application {
	return foundation.Setup().
		WithCommands(Commands).
		WithCommandsFilter(func() []string {
			return []string{
				"list",
			}
		}).
		WithProviders(Providers).
		WithConfig(config.Boot).
		Create()
}
