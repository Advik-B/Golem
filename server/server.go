package server

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"

	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/Advik-B/Golem/protocol/handshake"
	_ "github.com/Advik-B/Golem/protocol/login"
	_ "github.com/Advik-B/Golem/protocol/status"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

// Server implements the gnet.EventHandler interface.
type Server struct {
	gnet.BuiltinEventHandler // Correctly embed by value
	addr                     string

	// Server-wide state
	serverKey *rsa.PrivateKey
	motd      string
}

// NewServer creates a new Minecraft server instance.
func NewServer(addr string) *Server {
	// Generate a 2048-bit RSA key for encryption.
	// In a real server, this would be loaded from a file.
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Logger.Fatal("Failed to generate RSA server key", log.Error(err))
	}

	return &Server{
		addr:      addr,
		serverKey: key,
		motd:      `{"version":{"name":"Golem 1.21","protocol":767},"players":{"max":20,"online":0},"description":{"text":"A Golem Server"}}`,
	}
}

// Run starts the server and listens for connections.
func (s *Server) Run() error {
	log.Logger.Info("Starting Golem server...", zap.String("address", s.addr))
	frameCodec := &codec.VarIntFrameCodec{}
	return gnet.Run(s, s.addr, gnet.WithMulticore(true), gnet.WithCodec(frameCodec))
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Logger.Info("Server booted", zap.Int("event_loops", eng.NumEventLoop), zap.Bool("multicore", true))
	return
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	conn := &Connection{Conn: c}
	conn.SetState(protocol.Handshaking) // Set initial state
	c.SetContext(conn)
	log.Logger.Info("New connection", zap.Stringer("remote_addr", c.RemoteAddr()))
	return
}

func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		log.Logger.Warn("Connection closed with error", zap.Stringer("remote_addr", c.RemoteAddr()), zap.Error(err))
	} else {
		log.Logger.Info("Connection closed", zap.Stringer("remote_addr", c.RemoteAddr()))
	}
	return
}

// OnTraffic is where we handle incoming packets decoded by our FrameCodec.
func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	payload, err := c.Read()
	if err != nil {
		log.Logger.Error("Failed to read from connection", zap.Error(err), zap.Stringer("remote_addr", c.RemoteAddr()))
		return gnet.Close
	}

	conn, ok := c.Context().(*Connection)
	if !ok || conn == nil {
		log.Logger.Error("Context is nil or wrong type for connection, closing.", zap.Stringer("remote_addr", c.RemoteAddr()))
		return gnet.Close
	}

	packetBuf := codec.NewPacketBuffer(payload)
	if err := s.handlePacket(conn, packetBuf); err != nil {
		log.Logger.Error("Error handling packet", zap.Error(err), zap.Stringer("state", conn.State()), zap.Stringer("remote_addr", c.RemoteAddr()))
		return gnet.Close
	}

	return gnet.None
}

func (s *Server) handlePacket(conn *Connection, r *codec.PacketBuffer) error {
	packetID, err := r.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read packet ID: %w", err)
	}

	proto := protocol.GetProtocol(conn.State(), protocol.Serverbound)
	if proto == nil {
		return fmt.Errorf("no serverbound protocol defined for state %v", conn.State())
	}

	pk := proto.NewPacket(packetID)
	if pk == nil {
		log.Logger.Warn("Unknown packet", zap.Int32("id", packetID), zap.Stringer("state", conn.State()), zap.Int("size", r.Len()+1))
		return nil // Don't close connection for unknown packet, just skip
	}

	if err := pk.ReadFrom(r); err != nil {
		return fmt.Errorf("failed to read packet data for ID %#x (%T): %w", packetID, pk, err)
	}

	// --- Packet Dispatcher ---
	switch conn.State() {
	case protocol.Handshaking:
		return s.handleHandshake(conn, pk.(*handshake.ClientIntentionPacket))
	case protocol.Status:
		return s.handleStatus(conn, pk)
	case protocol.Login:
		return s.handleLogin(conn, pk)
	}

	return fmt.Errorf("handler for state %s not implemented", conn.State())
}

// SendPacket serializes and sends a packet to the client.
func (s *Server) SendPacket(c gnet.Conn, pk protocol.Packet) error {
	buf := codec.NewPacketBuffer(nil)
	buf.WriteVarInt(pk.ID())
	if err := pk.WriteTo(buf); err != nil {
		return fmt.Errorf("failed to write packet %T: %w", pk, err)
	}
	return c.AsyncWrite(buf.Bytes())
}
