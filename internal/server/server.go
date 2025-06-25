package server

import (
	"fmt"
	"net"
)

type Server struct {
	addr          string
	listener      net.Listener
	PlayerManager *PlayerManager
}

func NewServer(addr string) *Server {
	s := &Server{
		addr: addr,
	}
	s.PlayerManager = NewPlayerManager(s)
	return s
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("could not start listener: %w", err)
	}
	s.listener = listener
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Instead of handling the connection directly, we pass it to a pre-handler
		// that will eventually give it to the PlayerManager after the login is complete.
		go s.handleNewConnection(conn)
	}
}

// handleNewConnection now acts as a pre-login handler.
func (s *Server) handleNewConnection(conn net.Conn) {
	fmt.Printf("Accepted new connection from %s\n", conn.RemoteAddr())

	// Create a temporary connection object just for the handshake and login process.
	tempConn := NewConnection(s, conn)
	defer func() {
		// If the login process doesn't complete, this defer will close the raw connection.
		if tempConn.state != StatePlay {
			tempConn.Close()
		}
	}()

	// Handle only the handshake and login states.
	// The Play state will be handled by the PlayerManager's persistent connection.
	tempConn.HandleLogin()
}
