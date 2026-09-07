package providers

import (
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/defs"
	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/frpcclient"
	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/runners"
)

type FrppcServiceProvider struct {
	foundation.ServiceProviderWithRunners
}

func (r *FrppcServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{defs.FrppcRuntimeBinding},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

func (r *FrppcServiceProvider) Register(app foundation.Application) {
	app.Singleton(defs.FrppcRuntimeBinding, func(app foundation.Application) (any, error) {
		return frpcclient.NewServiceFromConfig()
	})
}

func (r *FrppcServiceProvider) Boot(app foundation.Application) {}

func (r *FrppcServiceProvider) Runners(app foundation.Application) []foundation.Runner {
	return runners.FrppcRunners()
}
