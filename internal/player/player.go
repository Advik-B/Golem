// internal/player/player.go

package player

import (
	"net"
)

type Player struct {
	Conn            net.Conn
	Username        string
	UUID            [16]byte
	State           int
	LastKeepAliveID int64
}
