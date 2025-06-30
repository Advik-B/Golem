package server

import (
	"sync"
	"time"

	"github.com/Advik-B/Golem/log"
	"github.com/Advik-B/Golem/protocol"
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
			zap.String("from", c.state.String()),
			zap.String("to", newState.String()),
		)
		c.state = newState
	}
}
