package wsgrpc

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
)

// GorillaConn wraps a gorilla/websocket.Conn as a net.Conn.
// Used by gRPC clients that dial over WebSocket (test commands, external clients).
type GorillaConn struct {
	ws         *gorillawebsocket.Conn
	readMutex  sync.Mutex
	writeMutex sync.Mutex
	readBuffer bytes.Buffer
}

// Compile-time guarantee that GorillaConn satisfies the net.Conn contract
// gRPC dials over.
var _ net.Conn = (*GorillaConn)(nil)

func NewGorillaConn(ws *gorillawebsocket.Conn) *GorillaConn {
	return &GorillaConn{ws: ws}
}

func (c *GorillaConn) Read(p []byte) (int, error) {
	c.readMutex.Lock()
	defer c.readMutex.Unlock()

	if c.readBuffer.Len() == 0 {
		messageType, data, err := c.ws.ReadMessage()
		if err != nil {
			return 0, err
		}
		if messageType != gorillawebsocket.BinaryMessage {
			return 0, fmt.Errorf("unexpected websocket message type: %d", messageType)
		}
		c.readBuffer.Write(data)
	}

	return c.readBuffer.Read(p)
}

func (c *GorillaConn) Write(p []byte) (int, error) {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	if err := c.ws.WriteMessage(gorillawebsocket.BinaryMessage, p); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (c *GorillaConn) Close() error {
	return c.ws.Close()
}

func (c *GorillaConn) LocalAddr() net.Addr {
	if conn := c.ws.UnderlyingConn(); conn != nil {
		return conn.LocalAddr()
	}

	return nil
}

func (c *GorillaConn) RemoteAddr() net.Addr {
	if conn := c.ws.UnderlyingConn(); conn != nil {
		return conn.RemoteAddr()
	}

	return nil
}

func (c *GorillaConn) SetDeadline(t time.Time) error {
	if err := c.ws.SetReadDeadline(t); err != nil {
		return err
	}

	return c.ws.SetWriteDeadline(t)
}

func (c *GorillaConn) SetReadDeadline(t time.Time) error {
	return c.ws.SetReadDeadline(t)
}

func (c *GorillaConn) SetWriteDeadline(t time.Time) error {
	return c.ws.SetWriteDeadline(t)
}
