// internal/server/server.go

package server

import (
	"log"
	"net" // Standard library net package
	"sync"
	"time"

	golemnet "github.com/Advik-B/Golem/internal/net"
	"github.com/Advik-B/Golem/internal/player"
)

type Server struct {
	listener net.Listener // This net.Listener is from the standard library
	players  map[net.Conn]*player.Player
	mu       sync.Mutex
}

func New(addr string) (*Server, error) {
	// net.Listen is from the standard library
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
		go s.handleConnection(conn)
	}
}

// AddPlayer safely adds a new player to the server's player map.
func (s *Server) AddPlayer(p *player.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[p.Conn] = p
	log.Printf("Player %s has joined the server.", p.Username)
}

// RemovePlayer safely removes a player from the server's player map.
func (s *Server) RemovePlayer(p *player.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, p.Conn)
	log.Printf("Player %s has left the server.", p.Username)
}

// handleConnection uses the aliased 'golemnet' package.
func (s *Server) handleConnection(c net.Conn) {
	defer c.Close()
	// Use the alias 'golemnet' to refer to your custom package
	connection := golemnet.NewConnection(c, s.AddPlayer)

	p, err := connection.Handle()
	if err != nil {
		if err.Error() != "Status packet handled" { // Don't log error for normal status pings
			log.Printf("Connection handler error for %s: %v", c.RemoteAddr(), err)
		}
		return
	}

	if p != nil {
		defer s.RemovePlayer(p)
		connection.HandlePlay()
	}
}

func (s *Server) startGameLoop() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for _, p := range s.players {
			if p.State == golemnet.PlayState {
				p.LastKeepAliveID = time.Now().Unix()
				w := golemnet.NewWriter(p.Conn)
				w.WritePacket(0x24, golemnet.WriteLong(p.LastKeepAliveID)) // Corrected Keep Alive Packet ID for 1.21 is 0x24
			}
		}
		s.mu.Unlock()
	}
}
