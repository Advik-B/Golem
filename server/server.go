package server

import (
	"go.uber.org/zap"
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
		return err
	}
	s.listener = listener
	defer s.listener.Close()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			Log.Error("Error accepting connection", zap.Error(err))
			continue
		}

		go s.handleNewConnection(conn)
	}
}

func (s *Server) handleNewConnection(conn net.Conn) {
	Log.Info("Accepted new connection", zap.String("addr", conn.RemoteAddr().String()))
	c := NewConnection(s, conn)
	c.Handle()
	s.PlayerManager.RemovePlayerByConn(c)
}
