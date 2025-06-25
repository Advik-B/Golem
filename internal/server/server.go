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

func (s *Server) handleNewConnection(conn net.Conn) {
	fmt.Printf("Accepted new connection from %s\n", conn.RemoteAddr())

	// Create ONE connection object to manage the client for its entire lifecycle.
	c := NewConnection(s, conn)

	// The Handle method will manage the connection from handshake to disconnect.
	// When it returns, the connection is over.
	c.Handle()

	// Clean up the player after the connection is closed.
	// This ensures the player is removed from the PlayerManager.
	s.PlayerManager.RemovePlayerByConn(c)
}
