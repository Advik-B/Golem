package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/panjf2000/gnet/v2"
	"go.uber.org/zap"
)

// Connection wraps a gnet.Conn and adds Minecraft-specific state management.
type Connection struct {
	gnet.Conn
	state      protocol.State
	stateMutex sync.RWMutex

	// Add other connection-specific data here
	Username  string
	Profile   protocol.GameProfile
	Latency   time.Duration
	Encrypted bool
	// ... etc
}

func (c *Connection) State() protocol.State {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

func (c *Connection) SetState(newState protocol.State) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	if c.state != newState {
		log.Logger.Debug("Connection state changed",
			zap.Stringer("remote_addr", c.RemoteAddr()),
			zap.Stringer("from", &c.state), // Pass as pointer to satisfy zap.Stringer
			zap.Stringer("to", &newState),
		)
		c.state = newState
	}
}

// WritePacket encodes a packet and writes it to the connection.
func (c *Connection) WritePacket(p protocol.Packet) error {
	// Create a buffer for the packet payload (ID + Data).
	payloadBuf := codec.NewPacketBuffer(nil)
	if err := payloadBuf.WriteVarInt(p.ID()); err != nil {
		return fmt.Errorf("failed to write packet ID: %w", err)
	}
	if err := p.WriteTo(payloadBuf); err != nil {
		return fmt.Errorf("failed to write packet data for %T: %w", p, err)
	}

	// gnet will call the codec's Encode method, which prepends the length.
	return c.AsyncWrite(payloadBuf.Bytes(), nil)
}
