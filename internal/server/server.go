package server

import (
	"bytes"
	"log"
	"net"
	"sync"
	"time"

	golemnet "github.com/Advik-B/Golem/internal/net"
	"github.com/Advik-B/Golem/internal/player"
	"github.com/Advik-B/Golem/nbt"
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
		go s.handleConnection(conn)
	}
}

func (s *Server) AddPlayer(p *player.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[p.Conn] = p
	log.Printf("Player %s has joined the server.", p.Username)
}

func (s *Server) RemovePlayer(p *player.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.players, p.Conn)
	log.Printf("Player %s has left the server.", p.Username)
}

func (s *Server) handleConnection(c net.Conn) {
	defer c.Close()
	connection := golemnet.NewConnection(c)

	// Phase 1: Handshake and Login. Returns a player object.
	p, err := connection.HandleLogin(s.AddPlayer)
	if err != nil {
		if err.Error() != "status packet handled" {
			log.Printf("Login failed for %s: %v", c.RemoteAddr(), err)
		}
		return
	}
	defer s.RemovePlayer(p)

	// Phase 2: Configuration.
	if err := connection.HandleConfiguration(); err != nil {
		log.Printf("Configuration failed for %s: %v", p.Username, err)
		return
	}

	// Phase 3: Transition to Play state.
	s.enterPlayState(p)
	connection.HandlePlay(p)
}

func (s *Server) startGameLoop() {
	ticker := time.NewTicker(time.Second * 15)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for _, p := range s.players {
			// REMOVED: Check for p.State. If a player is in this map,
			// they are in the play state and should get a keep-alive.
			p.LastKeepAliveID = time.Now().UnixMilli()
			w := golemnet.NewWriter(p.Conn)
			// ClientboundKeepAlivePacket for 1.21 is 0x24
			w.WritePacket(0x24, golemnet.WriteLong(p.LastKeepAliveID))
		}
		s.mu.Unlock()
	}
}

func (s *Server) enterPlayState(p *player.Player) {
	w := golemnet.NewWriter(p.Conn)

	// ClientboundLoginPacket (Join Game - 0x29)
	codecTag := golemnet.GetDimensionCodec()
	var codecBuf bytes.Buffer
	if err := nbt.Write(&codecBuf, nbt.NamedTag{Tag: codecTag}); err != nil {
		panic("failed to write dimension codec: " + err.Error())
	}

	w.WritePacket(0x29,
		golemnet.WriteInt(1), // entityId
		golemnet.WriteBool(false),
		golemnet.WriteVarInt(1),
		golemnet.WriteString("minecraft:overworld"),
		golemnet.WriteVarInt(20),
		golemnet.WriteVarInt(10),
		golemnet.WriteVarInt(10),
		golemnet.WriteBool(false),
		golemnet.WriteBool(true),
		golemnet.WriteBool(false),
		golemnet.WriteBool(true),
		golemnet.WriteBool(false),
		codecBuf.Bytes(),
		golemnet.WriteString("minecraft:overworld"),
		golemnet.WriteLong(0),
		golemnet.WriteByte(1),
		golemnet.WriteByte(0xFF),
		golemnet.WriteBool(false),
		golemnet.WriteBool(true),
		golemnet.WriteBool(false),
		golemnet.WriteVarInt(0),
	)

	// ClientboundCustomPayloadPacket (Brand) (0x19)
	brandPayload := golemnet.WriteString("Golem")
	w.WritePacket(0x19, golemnet.WriteString("minecraft:brand"), brandPayload)

	// ClientboundChangeDifficultyPacket (0x0E)
	w.WritePacket(0x0E, golemnet.WriteByte(1), golemnet.WriteBool(true))

	// ClientboundPlayerAbilitiesPacket (0x36)
	w.WritePacket(0x36, golemnet.WriteByte(0x06), golemnet.WriteFloat32(0.05), golemnet.WriteFloat32(0.1))

	// ClientboundSetHeldSlotPacket (0x51)
	w.WritePacket(0x51, golemnet.WriteByte(0))

	// ClientboundPlayerPositionPacket (0x40)
	w.WritePacket(0x40,
		golemnet.WriteDouble(0.0),
		golemnet.WriteDouble(100.0),
		golemnet.WriteDouble(0.0),
		golemnet.WriteFloat32(0.0),
		golemnet.WriteFloat32(0.0),
		golemnet.WriteByte(0),
		golemnet.WriteVarInt(1),
	)
}
