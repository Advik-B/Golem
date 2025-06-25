package server

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
)

type ConnectionState int

const (
	StateHandshaking   ConnectionState = 0
	StateStatus        ConnectionState = 1
	StateLogin         ConnectionState = 2
	StateConfiguration ConnectionState = 3 // NEW
	StatePlay          ConnectionState = 4 // Updated
)

type Connection struct {
	server *Server
	conn   net.Conn
	state  ConnectionState
	player *Player
}

func NewConnection(server *Server, conn net.Conn) *Connection {
	return &Connection{
		server: server,
		conn:   conn,
		state:  StateHandshaking,
	}
}

func (c *Connection) Close() {
	c.conn.Close()
}

func (c *Connection) Handle() {
	defer c.Close()

	for {
		packet, err := c.ReadPacket()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				// Don't log timeout errors, as they are common when a client pings and disconnects.
				// Any other read error is worth noting as a warning.
				var netErr net.Error
				if !errors.As(err, &netErr) || !netErr.Timeout() {
					Log.Warn("Error reading packet",
						zap.String("remoteAddr", c.conn.RemoteAddr().String()),
						zap.Error(err),
					)
				}
			}
			return
		}

		err = c.handlePacket(packet)
		if err != nil {
			// Check if it's our special "clean exit" signal for pings.
			if errors.Is(err, ErrPingComplete) {
				// This is expected. Break the loop and close the connection gracefully.
				return
			}

			// It's a real, unexpected error. Log it and close.
			Log.Error("Error handling packet",
				zap.String("remoteAddr", c.conn.RemoteAddr().String()),
				zap.Int32("packetID", packet.ID),
				zap.Error(err),
			)
			return
		}
	}
}

func (c *Connection) handlePacket(p Packet) error {
	switch c.state {
	case StateHandshaking:
		return c.handleHandshake(p)
	case StateStatus:
		return c.handleStatus(p)
	case StateLogin:
		return c.handleLogin(p)
	case StateConfiguration:
		return c.handleConfiguration(p)
	case StatePlay:
		return c.handlePlay(p)
	}
	return fmt.Errorf("unhandled state: %d", c.state)
}
