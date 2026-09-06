package frpcclient

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"

	"haokun-panel/cmd/frppc/facades"
	"haokun-panel/haokun/defs"
	"haokun-panel/utils/systemevent"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

type Service struct {
	opts options

	masterClient pb.ClientMasterClient
	conn         *grpc.ClientConn

	controller *clientController

	ctx    context.Context
	cancel context.CancelFunc

	streamLogMu     sync.Mutex
	streamLogCancel context.CancelFunc
	streamLogCh     chan string
	streamLogDone   chan struct{}

	initOnce     sync.Once
	shutdownOnce sync.Once

	eventDebouncerOnce sync.Once
	eventDebouncer     *systemevent.Debouncer[*pb.SystemEventPayload]
}

func NewServiceFromConfig() (*Service, error) {
	return NewServiceWithOptions(resolveOptionsFromConfig())
}

func NewServiceWithOptions(opts options) (*Service, error) {
	s := &Service{opts: opts, controller: NewClientController()}
	if !opts.enabled {
		return s, nil
	}
	if err := validateRequiredOptions(opts); err != nil {
		facades.Log().Warningf("frppc runtime disabled: %v", err)
		return s, nil
	}

	masterClient, conn, err := newMasterClient(opts.rpcURL)
	if err != nil {
		return nil, err
	}

	s.masterClient = masterClient
	s.conn = conn
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, nil
}

func (s *Service) ShouldRun() bool {
	return s != nil && s.opts.enabled && s.masterClient != nil && s.ctx != nil && s.cancel != nil
}

func (s *Service) Context() context.Context {
	if s == nil || s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

func (s *Service) SyncInterval() time.Duration {
	if s == nil || s.opts.syncInterval <= 0 {
		return 30 * time.Second
	}
	return s.opts.syncInterval
}

func (s *Service) EnsureInitialized() {
	if !s.ShouldRun() {
		return
	}
	s.initOnce.Do(func() {
		facades.Log().Infof("starting frp client runtime id=%s rpc-url=%s", s.opts.clientID, s.opts.rpcURL)
		if err := s.PullConfig(s.ctx); err != nil {
			facades.Log().Warningf("initial pull client config failed, will retry: %v", err)
		}
	})
}

func (s *Service) Shutdown() error {
	if s == nil {
		return nil
	}

	var closeErr error
	s.shutdownOnce.Do(func() {
		// Report the stop before tearing down the connection, while the master
		// stream is still usable. Best-effort — never blocks shutdown. Stop()
		// flushes any pending debounced aggregate synchronously so nothing is
		// lost when the connection closes below.
		s.emitLifecycleEvent(defs.SystemEventTypeFrppcStopped, defs.SystemEventLevelWarn)
		s.events().Stop()
		if s.cancel != nil {
			s.cancel()
		}
		if err := s.stopStreamLog(); err != nil && closeErr == nil {
			closeErr = err
		}
		if s.controller != nil {
			s.controller.StopAll()
			s.controller.DeleteAll()
		}
		if s.conn != nil {
			if err := s.conn.Close(); err != nil && closeErr == nil {
				closeErr = err
			}
		}
	})
	return closeErr
}
