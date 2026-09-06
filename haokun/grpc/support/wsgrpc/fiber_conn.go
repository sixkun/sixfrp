package wsgrpc

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	fiberws "github.com/gofiber/contrib/v3/websocket"
)

// wsAddr is a net.Addr carrying the real client address recovered from the
// websocket handshake (e.g. X-Forwarded-For). We only have an IP, not a port,
// so String() returns the bare IP. It exists so gRPC's peer.FromContext sees
// the true client instead of the reverse proxy that terminated the TCP hop.
type wsAddr struct {
	addr string
}

func (a wsAddr) Network() string { return "ws" }
func (a wsAddr) String() string  { return a.addr }

// Compile-time guarantees: FiberConn must satisfy net.Conn (gRPC serves it via
// the ws Listener) and wsAddr must satisfy net.Addr (returned by RemoteAddr).
var (
	_ net.Conn = (*FiberConn)(nil)
	_ net.Addr = wsAddr{}
)

type FiberConn struct {
	ws         *fiberws.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	readBuffer bytes.Buffer
	done       chan struct{}
	closeOnce  sync.Once

	// remoteAddr, when non-nil, overrides the underlying TCP peer. Behind a
	// reverse proxy the underlying conn only sees the proxy, so the gateway
	// sets this from the handshake's forwarded client IP.
	remoteAddr net.Addr
}

func NewFiberConn(ws *fiberws.Conn) *FiberConn {
	return &FiberConn{
		ws:   ws,
		done: make(chan struct{}),
	}
}

// SetRemoteAddr overrides the address reported by RemoteAddr. Pass the real
// client IP recovered from the websocket handshake (e.g. X-Forwarded-For).
// A blank ip is ignored so the underlying TCP peer stays in effect.
func (c *FiberConn) SetRemoteAddr(ip string) {
	if ip == "" {
		return
	}
	c.remoteAddr = wsAddr{addr: ip}
}

func (c *FiberConn) Done() <-chan struct{} {
	return c.done
}

func (c *FiberConn) Read(p []byte) (int, error) {
	c.readMutex.Lock()
	defer c.readMutex.Unlock()

	if c.readBuffer.Len() == 0 {
		messageType, data, err := c.ws.ReadMessage()
		if err != nil {
			return 0, err
		}
		if messageType != fiberws.BinaryMessage {
			return 0, fmt.Errorf("unexpected websocket message type: %d", messageType)
		}
		c.readBuffer.Write(data)
	}

	return c.readBuffer.Read(p)
}

func (c *FiberConn) Write(p []byte) (int, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if err := c.ws.WriteMessage(fiberws.BinaryMessage, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (c *FiberConn) Close() error {
	var err error
	c.closeOnce.Do(func() {
		err = c.ws.Close()
		close(c.done)
	})
	return err
}

func (c *FiberConn) LocalAddr() net.Addr {
	if conn := c.ws.NetConn(); conn != nil {
		return conn.LocalAddr()
	}

	return nil
}

func (c *FiberConn) RemoteAddr() net.Addr {
	if c.remoteAddr != nil {
		return c.remoteAddr
	}
	if conn := c.ws.NetConn(); conn != nil {
		return conn.RemoteAddr()
	}

	return nil
}

func (c *FiberConn) SetDeadline(t time.Time) error {
	if err := c.ws.SetReadDeadline(t); err != nil {
		return err
	}

	return c.ws.SetWriteDeadline(t)
}

func (c *FiberConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *FiberConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}
