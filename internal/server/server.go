package server

import (
	inet "github.com/Advik-B/Golem/internal/net" // Corrected import
	"github.com/Advik-B/Golem/internal/player"   // Corrected import
	"log"
	"net"
	"sync"
	"time"
)

type Server struct {
	listener net.Listener
	players  map[net.Conn]*player.Player
	mu       sync.Mutex
}

func New(addr string) (*Server, error) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: ln,
		players:  make(map[net.Conn]*player.Player),
	}, nil
}

func (s *Server) Start() {
	log.Println("Golem server listening on", s.listener.Addr())
	go s.startGameLoop()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		log.Println("Accepted new connection from", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(c net.Conn) {
	connection := inet.NewConnection(c)
	connection.Handle()
	// When Handle() returns, the connection is closed.
	// We'll add player removal logic here later.
	log.Println("Connection closed for", c.RemoteAddr())
}

func (s *Server) startGameLoop() {
	ticker := time.NewTicker(50 * time.Millisecond) // 20 ticks per second
	defer ticker.Stop()

	for range ticker.C {
		// This is the main game tick.
		// We will add logic here later, such as:
		// - Sending keep-alive packets
		// - Ticking entities
		// - Processing player movement
	}
}
