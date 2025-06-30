package server

import (
	"fmt"
	"github.com/Advik-B/Golem/protocol"
	"time"

	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

// Server implements the gnet.EventHandler interface.
type Server struct {
	*gnet.BuiltinEventHandler
	addr string
	// Add server-wide state here, e.g., a player list
}

// NewServer creates a new Minecraft server instance.
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// Run starts the server and listens for connections.
func (s *Server) Run() error {
	log.Logger.Info("Starting Golem server...", zap.String("address", s.addr))
	// We need a codec to handle Minecraft's length-prefixed framing.
	codec := &codec.VarIntFrameCodec{}
	return gnet.Run(s, s.addr, gnet.WithMulticore(true), gnet.WithCodec(codec))
}

// OnBoot is called when the server starts.
func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	log.Logger.Info("Server booted",
		zap.Int("goroutine_count", eng.NumEventLoop),
		zap.Bool("multicore", true),
	)
	return
}

// OnOpen is called when a new connection is established.
func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	// Wrap the gnet.Conn in our custom Connection struct
	conn := &Connection{Conn: c}
	c.SetContext(conn)
	log.Logger.Info("New connection", zap.Stringer("remote_addr", c.RemoteAddr()))
	return
}

// OnClose is called when a connection is closed.
func (s *Server) OnClose(c gnet.Conn, err error) (action gnet.Action) {
	if err != nil {
		log.Logger.Warn("Connection closed with error", zap.Stringer("remote_addr", c.RemoteAddr()), zap.Error(err))
	} else {
		log.Logger.Info("Connection closed", zap.Stringer("remote_addr", c.RemoteAddr()))
	}
	// Here you would handle player disconnection logic
	return
}

// OnTick is called periodically.
func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	// Main server game loop could be ticked here
	delay = 1 * time.Second // For now, just a slow tick
	return
}

// OnTraffic is where we handle incoming packets.
func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	ctx := c.Context()
	if ctx == nil {
		log.Logger.Error("Context is nil for connection, closing.", zap.Stringer("remote_addr", c.RemoteAddr()))
		return gnet.Close
	}

	conn := ctx.(*Connection)
	buf, err := c.Read()
	if err != nil {
		log.Logger.Error("Failed to read from connection", zap.Error(err))
		return gnet.Close
	}
	if len(buf) == 0 {
		return gnet.None
	}

	// Create a PacketBuffer to read from the received data
	packetBuf := codec.NewPacketBuffer(buf)
	for packetBuf.Len() > 0 {
		if err := s.handlePacket(conn, packetBuf); err != nil {
			log.Logger.Error("Error handling packet", zap.Error(err), zap.Stringer("state", conn.State()))
			return gnet.Close
		}
	}
	return gnet.None
}

func (s *Server) handlePacket(conn *Connection, r *codec.PacketBuffer) error {
	// The VarIntFrameCodec has already read the length. We just need the ID.
	packetID, err := r.ReadVarInt()
	if err != nil {
		return fmt.Errorf("failed to read packet ID: %w", err)
	}

	proto := protocol.GetProtocol(conn.State(), protocol.Serverbound)
	if proto == nil {
		return fmt.Errorf("no protocol defined for state %v", conn.State())
	}

	pk := proto.NewPacket(packetID)
	if pk == nil {
		return fmt.Errorf("unknown packet ID %#x in state %v", packetID, conn.State())
	}

	if err := pk.ReadFrom(r); err != nil {
		return fmt.Errorf("failed to read packet data for ID %#x: %w", packetID, err)
	}

	// TODO: Dispatch packet to the correct handler based on state.
	log.Logger.Debug("Received packet",
		zap.Stringer("state", conn.State()),
		zap.Int32("id", packetID),
		zap.String("type", fmt.Sprintf("%T", pk)),
	)
	return nil
}
