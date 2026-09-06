package wsgrpc

import (
	"fmt"
	"net"
	"sync"
)

type Listener struct {
	connCh chan net.Conn
	done   chan struct{}
	mu     sync.Mutex
	closed bool
	addr   net.Addr
}

/**
 * extends net.Listener
 */
type listenerAddr struct {
	net.Listener
	network string
	address string
}

func (a listenerAddr) Network() string {
	return a.network
}

func (a listenerAddr) String() string {
	return a.address
}

func NewListener(address, network string, bufferSize int) *Listener {
	return &Listener{
		connCh: make(chan net.Conn, bufferSize),
		done:   make(chan struct{}),
		addr:   listenerAddr{network: network, address: address},
	}
}

func (l *Listener) Accept() (net.Conn, error) {
	select {
	case conn, ok := <-l.connCh:
		if !ok {
			return nil, fmt.Errorf("grpc websocket listener closed")
		}
		return conn, nil
	case <-l.done:
		return nil, fmt.Errorf("grpc websocket listener closed")
	}
}

func (l *Listener) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}

	l.closed = true
	close(l.done)

	// Drain and close any queued connections that will never be accepted
	for {
		select {
		case conn := <-l.connCh:
			if conn != nil {
				_ = conn.Close()
			}
		default:
			close(l.connCh)
			return nil
		}
	}
}

// Done returns a channel that is closed when the listener is shut down.
// Use this in select statements to detect shutdown and unblock waiting goroutines.
func (l *Listener) Done() <-chan struct{} {
	return l.done
}

func (l *Listener) Addr() net.Addr {
	return l.addr
}

// Enqueue hands a connection to the gRPC server via the listener.
// Returns an error if the listener is closed or the queue is full.
func (l *Listener) Enqueue(conn net.Conn) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return fmt.Errorf("grpc websocket listener closed")
	}

	select {
	case l.connCh <- conn:
		return nil
	default:
		return fmt.Errorf("grpc websocket listener queue is full")
	}
}
