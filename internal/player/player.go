package player

import (
	"net"
)

// Player represents a player in the game world. It no longer tracks network state.
type Player struct {
	Conn            net.Conn
	Username        string
	UUID            [16]byte
	LastKeepAliveID int64
}

// New creates a new player instance.
func New(conn net.Conn, username string) *Player {
	return &Player{
		Conn:     conn,
		Username: username,
	}
}
