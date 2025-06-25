package server

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

type Player struct {
	EntityID int32
	UUID     uuid.UUID
	Username string
	Conn     *Connection
}

type PlayerManager struct {
	players map[uuid.UUID]*Player
	server  *Server
	sync.RWMutex
}

func NewPlayerManager(s *Server) *PlayerManager {
	return &PlayerManager{
		players: make(map[uuid.UUID]*Player),
		server:  s,
	}
}

func (pm *PlayerManager) AddPlayer(conn net.Conn, username string, playerUUID uuid.UUID) *Player {
	pm.Lock()
	defer pm.Unlock()

	entityID := int32(len(pm.players) + 1)

	player := &Player{
		EntityID: entityID,
		UUID:     playerUUID,
		Username: username,
		Conn:     NewConnection(pm.server, conn),
	}

	pm.players[playerUUID] = player

	go pm.handlePlayerConnection(player)

	return player
}

func (pm *PlayerManager) handlePlayerConnection(player *Player) {
	defer pm.RemovePlayer(player)
	player.Conn.Handle()
}

func (pm *PlayerManager) RemovePlayer(player *Player) {
	pm.Lock()
	defer pm.Unlock()

	delete(pm.players, player.UUID)
	player.Conn.Close()
	fmt.Printf("Player %s disconnected.\n", player.Username)
}

// GetPlayerByConn finds a player associated with a specific connection object.
// THIS IS THE NEWLY ADDED FUNCTION.
func (pm *PlayerManager) GetPlayerByConn(c *Connection) *Player {
	pm.RLock()
	defer pm.RUnlock()
	for _, player := range pm.players {
		if player.Conn == c {
			return player
		}
	}
	return nil
}
