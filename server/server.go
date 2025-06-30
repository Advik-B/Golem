package server

import (
	"fmt"
	"time"

	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

// Server implements the gnet.EventHandler interface.
type Server struct {
	*gnet.BuiltinEventEngine // Corrected: Embed the struct from the documentation
	addr                     string
	frameCodec               *codec.VarIntFrameCodec // The codec is stateless, but keeping it on the struct is fine.
}

// NewServer creates a new Minecraft server instance.
func NewServer(addr string) *Server {
	return &Server{
		addr:       addr,
		frameCodec: &codec.VarIntFrameCodec{},
	}
}

// Run starts the server and listens for connections.
func (s *Server) Run() error {
	log.Logger.Info("Starting Golem server...", zap.String("address", s.addr))
	// Corrected: WithCodec is not a valid option in the provided documentation.
	return gnet.Run(s, s.addr, gnet.WithMulticore(true))
}

func (s *Server) OnBoot(eng gnet.Engine) (action gnet.Action) {
	// Corrected: NumEventLoop() does not exist on the Engine object per the documentation.
	log.Logger.Info("Server booted", zap.Bool("multicore", true))
	return
}

func (s *Server) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	conn := &Connection{Conn: c}
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

func (s *Server) OnTick() (delay time.Duration, action gnet.Action) {
	delay = 2 * time.Second
	return
}

// OnTraffic now implements manual framing. It's called when there is data in the connection's inbound buffer.
func (s *Server) OnTraffic(c gnet.Conn) (action gnet.Action) {
	conn, ok := c.Context().(*Connection)
	if !ok || conn == nil {
		log.Logger.Error("Context is nil or wrong type for connection, closing.", zap.Stringer("remote_addr", c.RemoteAddr()))
		return gnet.Close
	}

	for {
		// Use the frame codec to decode one full packet frame from the connection's stream.
		// The codec's Decode method now takes the gnet.Conn directly.
		frame, err := s.frameCodec.Decode(c)
		if err != nil {
			log.Logger.Error("Frame decode error", zap.Error(err), zap.Stringer("remote_addr", c.RemoteAddr()))
			return gnet.Close
		}
		if frame == nil {
			// Not enough data for a full frame, break the loop and wait for more network traffic.
			break
		}

		// We have a full frame, process it.
		packetBuf := codec.NewPacketBuffer(frame)
		if err := s.handlePacket(conn, packetBuf); err != nil {
			log.Logger.Error("Error handling packet", zap.Error(err), zap.Stringer("state", &conn.state), zap.Stringer("remote_addr", c.RemoteAddr()))
			return gnet.Close
		}
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
		log.Logger.Warn("Unknown packet", zap.Int32("id", packetID), zap.Stringer("state", &conn.state), zap.Int("size", r.Len()+1))
		return nil
	}

	if err := pk.ReadFrom(r); err != nil {
		return fmt.Errorf("failed to read packet data for ID %#x (%T): %w", packetID, pk, err)
	}

	log.Logger.Debug("Received packet",
		zap.Stringer("state", &conn.state),
		zap.Int32("id", packetID),
		zap.String("type", fmt.Sprintf("%T", pk)),
	)
	// TODO: Dispatch packet to the correct handler based on state.
	return nil
}
