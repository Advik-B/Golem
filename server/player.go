package server

import (
	"fmt"
	"github.com/google/uuid"
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

func (pm *PlayerManager) AddPlayer(c *Connection, username string, playerUUID uuid.UUID) *Player {
	pm.Lock()
	defer pm.Unlock()

	entityID := int32(len(pm.players) + 1) // Simple entity ID for now

	player := &Player{
		EntityID: entityID,
		UUID:     playerUUID,
		Username: username,
		Conn:     c, // Use the existing connection
	}

	// Associate the player with the connection
	c.player = player
	pm.players[playerUUID] = player

	fmt.Printf("Player %s [%s] added to PlayerManager.\n", username, playerUUID)

	return player
}

func (pm *PlayerManager) handlePlayerConnection(player *Player) {
	defer pm.RemovePlayer(player)
	player.Conn.Handle()
}

func (pm *PlayerManager) RemovePlayer(player *Player) {
	if player == nil {
		return
	}
	pm.Lock()
	defer pm.Unlock()

	if _, ok := pm.players[player.UUID]; ok {
		delete(pm.players, player.UUID)
		player.Conn.Close()
		fmt.Printf("Player %s disconnected.\n", player.Username)
	}
}

// GetPlayerByConn finds a player associated with a specific connection object.
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

func (pm *PlayerManager) RemovePlayerByConn(c *Connection) {
	// The connection might have closed before a player was ever created.
	if c.player != nil {
		pm.RemovePlayer(c.player)
	} else {
		fmt.Printf("Connection from %s closed before login.\n", c.conn.RemoteAddr())
	}
}
