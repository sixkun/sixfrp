package contracts

import (
	"context"
	"time"
)

type Service interface {
	ShouldRun() bool
	EnsureInitialized()
	Context() context.Context
	SyncInterval() time.Duration

	PullConfig(ctx context.Context) error
	ConnectAndServeRPC(ctx context.Context) error

	Shutdown() error
}
