// Package systemevent provides a sender-side debouncer for system events.
//
// It implements an RxJS-style debounce with a LEADING emit plus a debounced
// TRAILING aggregate, keyed by an event's dedupe_key:
//
//   - The first event for a key is emitted immediately (leading edge).
//   - Subsequent same-key events within the silence window are buffered, not
//     emitted. Each new same-key event resets the window (true debounce).
//   - After the window elapses with no further same-key event, one aggregated
//     trailing event is emitted carrying the count of buffered occurrences.
//
// Events with an empty dedupe_key bypass the debouncer and are emitted
// immediately and individually. The debouncer is transport-agnostic — the caller
// supplies both the payload type and the emit callback — so one implementation
// serves master's in-process dispatch and an agent's outbound gRPC pushes.
package systemevent

import (
	"sync"
	"time"
)

// Payload is the shape the debouncer needs from an event: the dedupe key it
// groups by, and nothing else.
//
// Deliberately NOT proto.Message. The agents debounce their own outbound wire
// messages (each generated from its own contract), but master debounces its own
// connectivity events, which never leave the process — requiring a proto message
// only forced master to borrow a wire type it had no use for.
//
// comparable is what makes Submit's "did I get a zero value" guard a compile-time
// safe `event == zero` instead of an `any(event) == any(zero)` that panics on a
// non-comparable dynamic type. Pointer types satisfy it, which is what every
// caller uses.
//
// The aggregate-count annotation is deliberately not here either — it has to
// WRITE a field, and doing that through a type parameter means reflection. Each
// side owns a concrete wrapper over MergeAggregateCount instead.
type Payload interface {
	comparable
	GetDedupeKey() string
}

// EmitFunc delivers a (possibly aggregated) event to its transport. It is
// called from the debouncer's own goroutine for trailing emits and inline for
// leading emits, so it must be safe to call concurrently and should not block
// for long.
type EmitFunc[T Payload] func(event T, count int)

// Debouncer coalesces storming same-key events. A zero Debouncer is not usable;
// construct with New. It is safe for concurrent use.
type Debouncer[T Payload] struct {
	window time.Duration
	emit   EmitFunc[T]

	mu      sync.Mutex
	pending map[string]*pendingKey[T]
	stopped bool
}

type pendingKey[T Payload] struct {
	// last is the most recent event seen for this key; it is the template for
	// the trailing aggregate emit.
	last T
	// count is how many same-key events have been buffered since the leading
	// emit (i.e. excluding the leading one).
	count int
	timer *time.Timer
}

// New returns a Debouncer that emits through emit using the given silence
// window. A window <= 0 disables debouncing (every event emits immediately).
func New[T Payload](window time.Duration, emit EmitFunc[T]) *Debouncer[T] {
	return &Debouncer[T]{
		window:  window,
		emit:    emit,
		pending: make(map[string]*pendingKey[T]),
	}
}

// Submit feeds an event into the debouncer. Events without a dedupe_key (or
// when the window is disabled) emit immediately. The first event for a key
// emits immediately (leading); further same-key events are buffered and later
// collapsed into one trailing aggregate.
func (d *Debouncer[T]) Submit(event T) {
	var zero T
	if event == zero {
		return
	}
	key := event.GetDedupeKey()
	if key == "" || d.window <= 0 {
		d.emit(event, 1)
		return
	}

	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		d.emit(event, 1)
		return
	}

	if p, ok := d.pending[key]; ok {
		// Within an active window: buffer, reset the timer (debounce), and hold.
		p.last = event
		p.count++
		p.timer.Reset(d.window)
		d.mu.Unlock()
		return
	}

	// Leading edge: no active window for this key. Emit now and open a window
	// so that any follow-ups are aggregated into a single trailing event.
	d.pending[key] = &pendingKey[T]{
		last:  event,
		count: 0,
		timer: time.AfterFunc(d.window, func() { d.flush(key) }),
	}
	d.mu.Unlock()

	d.emit(event, 1)
}

// flush emits the trailing aggregate for key (if any follow-ups were buffered)
// and clears its pending state. Invoked by the key's timer.
func (d *Debouncer[T]) flush(key string) {
	d.mu.Lock()
	p, ok := d.pending[key]
	if !ok {
		d.mu.Unlock()
		return
	}
	delete(d.pending, key)
	d.mu.Unlock()

	// No follow-ups after the leading emit → nothing to aggregate.
	if p.count == 0 {
		return
	}
	d.emit(p.last, p.count)
}

// Stop cancels all pending timers and flushes any buffered aggregates
// immediately, so nothing is lost on shutdown. After Stop the debouncer emits
// every Submit inline.
func (d *Debouncer[T]) Stop() {
	d.mu.Lock()
	if d.stopped {
		d.mu.Unlock()
		return
	}
	d.stopped = true
	flushes := make([]*pendingKey[T], 0, len(d.pending))
	for key, p := range d.pending {
		p.timer.Stop()
		flushes = append(flushes, p)
		delete(d.pending, key)
	}
	d.mu.Unlock()

	for _, p := range flushes {
		if p.count > 0 {
			d.emit(p.last, p.count)
		}
	}
}
