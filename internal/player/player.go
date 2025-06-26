package player

import (
	"net"
)

type Player struct {
	Conn     net.Conn
	Username string
	UUID     [16]byte
	// Add other player-related fields here later (e.g., position, health)
}
