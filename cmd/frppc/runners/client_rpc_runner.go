package runners

import (
	"context"
	"errors"
	"io"
	"time"

	"haokun-panel/cmd/frppc/contracts"
	"haokun-panel/cmd/frppc/facades"
)

type ClientRPCRunner struct {
	service contracts.Service
	cancel  context.CancelFunc
}

func NewClientRPCRunner() *ClientRPCRunner {
	return &ClientRPCRunner{service: facades.FrpClientService()}
}

func (r *ClientRPCRunner) Signature() string {
	return "frppc-client-rpc-stream"
}

func (r *ClientRPCRunner) ShouldRun() bool {
	return r.service != nil && r.service.ShouldRun()
}

func (r *ClientRPCRunner) Run() error {
	r.service.EnsureInitialized()

	ctx, cancel := context.WithCancel(r.service.Context())
	r.cancel = cancel

	const reconnectDelay = 3 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		err := r.service.ConnectAndServeRPC(ctx)
		if ctx.Err() != nil {
			return nil
		}
		if err == nil {
			err = io.EOF
		}

		// EOF/Canceled can also come from a remote stream closure. Only the
		// runner's own canceled context is a terminal condition.
		if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
			facades.Log().Warningf("rpc stream closed: %v, reconnecting in %s", err, reconnectDelay)
		} else {
			facades.Log().Warningf("rpc stream disconnected: %v, reconnecting in %s", err, reconnectDelay)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(reconnectDelay):
		}
	}
}

func (r *ClientRPCRunner) Shutdown() error {
	if r.cancel != nil {
		r.cancel()
	}
	return r.service.Shutdown()
}
