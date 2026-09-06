package facades

import (
	"haokun-panel/cmd/frppc/contracts"
	"haokun-panel/cmd/frppc/defs"
)

func FrpClientService() contracts.Service {
	instance, err := App().Make(defs.FrppcRuntimeBinding)
	if err != nil {
		Log().Fatalf("resolve frppc runtime failed: %v", err)
		return nil
	}
	binding, ok := instance.(contracts.Service)
	if !ok {
		Log().Fatalf("instance is not a frppc contracts.Service")
		return nil
	}
	return binding
}
