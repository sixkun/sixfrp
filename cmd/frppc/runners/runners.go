package runners

import (
	"github.com/goravel/framework/contracts/foundation"
)

func FrppcRunners() []foundation.Runner {
	return []foundation.Runner{
		NewClientRPCRunner(),
		NewPeriodicPullRunner(),
	}
}
