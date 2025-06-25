package server

import (
	"errors"
	"fmt"
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
			if err != io.EOF {
				// Don't log timeout errors during ping, as they are expected
				// if the client disconnects first. Any other read error is a problem.
				if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
					fmt.Printf("Error reading packet from %s: %v\n", c.conn.RemoteAddr(), err)
				}
			}
			return
		}

		err = c.handlePacket(packet)
		if err != nil {
			// Check if it's our special "clean exit" signal.
			if errors.Is(err, ErrPingComplete) {
				// This is expected. Break the loop and close the connection gracefully.
				return
			}

			// It's a real, unexpected error. Log it and close.
			fmt.Printf("Error handling packet from %s: %v\n", c.conn.RemoteAddr(), err)
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
