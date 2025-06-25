package server

import (
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

// HandleLogin processes packets up until the login sequence is complete and transitions to configuration.
func (c *Connection) HandleLogin() {
	// Loop until we are no longer in Handshaking or Status. We should end in the Login state.
	for c.state == StateHandshaking || c.state == StateStatus {
		packet, err := c.ReadPacket()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading packet from %s during pre-login: %v\n", c.conn.RemoteAddr(), err)
			}
			return
		}
		err = c.handlePacket(packet)
		if err != nil {
			fmt.Printf("Error handling packet from %s during pre-login: %v\n", c.conn.RemoteAddr(), err)
			return
		}
	}

	// Now that we are in the Login state, we handle the final login packet.
	// This will transition the player to the PlayerManager and the Configuration state.
	loginPacket, err := c.ReadPacket()
	if err != nil {
		fmt.Printf("Error reading login start packet: %v\n", err)
		return
	}
	_ = c.handleLogin(loginPacket)
}

// Handle is the main loop for a player that is already in the game (or in configuration).
func (c *Connection) Handle() {
	for {
		packet, err := c.ReadPacket()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading packet from %s: %v\n", c.conn.RemoteAddr(), err)
			}
			return
		}

		err = c.handlePacket(packet)
		if err != nil {
			fmt.Printf("Error handling packet from %s: %v\n", c.conn.RemoteAddr(), err)
			return
		}

		// If we've successfully transitioned out of Play, the connection is over.
		if c.state != StatePlay && c.state != StateConfiguration {
			break
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
