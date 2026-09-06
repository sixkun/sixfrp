package runners

import (
	"context"
	"time"

	"haokun-panel/cmd/frppc/contracts"
	"haokun-panel/cmd/frppc/facades"
)

type PeriodicPullRunner struct {
	service contracts.Service
	cancel  context.CancelFunc
}

func NewPeriodicPullRunner() *PeriodicPullRunner {
	return &PeriodicPullRunner{service: facades.FrpClientService()}
}

func (r *PeriodicPullRunner) Signature() string {
	return "frppc-periodic-pull"
}

func (r *PeriodicPullRunner) ShouldRun() bool {
	return r.service != nil && r.service.ShouldRun()
}

func (r *PeriodicPullRunner) Run() error {
	r.service.EnsureInitialized()

	ctx, cancel := context.WithCancel(r.service.Context())
	r.cancel = cancel

	interval := r.service.SyncInterval()
	if interval <= 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := r.service.PullConfig(ctx); err != nil {
				facades.Log().Warningf("periodic pull client config failed: %v", err)
			}
		}
	}
}

func (r *PeriodicPullRunner) Shutdown() error {
	if r.cancel != nil {
		r.cancel()
	}
	return nil
}
