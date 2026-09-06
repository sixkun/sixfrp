// Package streamlog forwards goravel log entries into the frppc stream-log
// channel so that records emitted via facades.Log() while a stream session is
// active are pushed to master alongside the existing stream-log pump.
package streamlog

import (
	"fmt"
	"strings"
	"sync"
	"time"

	contractslog "github.com/goravel/framework/contracts/log"
)

// Sink forwards log entries to the single channel attached by the frppc
// service while a stream session is active. Multi-subscriber fan-out lives in
// the browser client, not here.
type Sink struct {
	mu  sync.RWMutex
	ch  chan<- string
	lvl contractslog.Level
}

// Default is the process-wide sink wired into the custom log channel.
var Default = &Sink{lvl: contractslog.LevelDebug}

func (s *Sink) Attach(ch chan<- string) {
	s.mu.Lock()
	s.ch = ch
	s.mu.Unlock()
}

func (s *Sink) Detach() {
	s.mu.Lock()
	s.ch = nil
	s.mu.Unlock()
}

func (s *Sink) SetLevel(l contractslog.Level) {
	s.mu.Lock()
	s.lvl = l
	s.mu.Unlock()
}

// NewLogger returns the contracts/log.Logger adapter that exposes s to the
// goravel custom-driver wiring (see config "via" entry).
func NewLogger(s *Sink) contractslog.Logger {
	return &logger{s: s}
}

type logger struct {
	s *Sink
}

func (l *logger) Handle(channel string) (contractslog.Handler, error) {
	return &handler{s: l.s}, nil
}

type handler struct {
	s *Sink
}

func (h *handler) Enabled(level contractslog.Level) bool {
	h.s.mu.RLock()
	defer h.s.mu.RUnlock()
	return level >= h.s.lvl
}

func (h *handler) Handle(entry contractslog.Entry) error {
	h.s.mu.RLock()
	ch := h.s.ch
	h.s.mu.RUnlock()
	if ch == nil {
		return nil
	}

	select {
	case ch <- formatEntry(entry):
	default:
	}
	return nil
}

func formatEntry(e contractslog.Entry) string {
	ts := e.Time()
	if ts.IsZero() {
		ts = time.Now()
	}
	msg := strings.TrimRight(e.Message(), "\n")
	return fmt.Sprintf("%s [%s] %s", ts.Format(time.RFC3339), e.Level().String(), msg)
}
