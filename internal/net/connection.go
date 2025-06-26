package net

import (
	"bytes"
	"github.com/Advik-B/Golem/internal/player"
	"io"
	"log"
	"net"
)

const (
	HandshakeState = 0
	StatusState    = 1
	LoginState     = 2
	PlayState      = 3
)

type Connection struct {
	conn  net.Conn
	state int
	r     *Reader
	w     *Writer
	p     *player.Player
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn:  conn,
		state: HandshakeState,
		r:     NewReader(conn),
		w:     NewWriter(conn),
	}
}

func (c *Connection) Handle() {
	defer c.conn.Close()
	for {
		// This is the main packet-reading loop for the connection
		pktLen, err := c.r.ReadVarInt()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading packet length: %v", err)
			}
			return
		}

		data := make([]byte, pktLen)
		if _, err := c.r.Read(data); err != nil {
			log.Printf("Error reading packet data: %v", err)
			return
		}

		packetReader := NewReader(bytes.NewReader(data))
		pktID, _ := packetReader.ReadVarInt()

		switch c.state {
		case HandshakeState:
			// The only packet here is Handshake (0x00)
			c.handleHandshake(packetReader)
		case StatusState:
			c.handleStatus(pktID, packetReader)
		case LoginState:
			c.handleLogin(pktID, packetReader)
		case PlayState:
			// Once in play state, we'll need a different loop, likely involving channels.
			// For now, we can just log that we are in the play state and stop handling.
			log.Println("Reached play state. Packet handling will be implemented later.")
			return
		}
	}
}

func (c *Connection) handleHandshake(r *Reader) {
	_, _ = r.ReadVarInt() // protocol version
	_, _ = r.ReadString() // server address
	r.r.ReadByte()        // port high byte
	r.r.ReadByte()        // port low byte
	nextState, _ := r.ReadVarInt()
	c.state = nextState
}

func (c *Connection) handleStatus(pktID int, r *Reader) {
	if pktID == 0x00 { // Status Request
		resp := `{"version":{"name":"Golem 1.21.5","protocol":765},"players":{"max":100,"online":0,"sample":[]},"description":{"text":"A Golem Server 🗿"}}`
		c.w.WritePacket(0x00, WriteString(resp))
	} else if pktID == 0x01 { // Ping Request
		payload := make([]byte, 8)
		r.Read(payload)
		c.w.WritePacket(0x01, payload)
	}
}

func (c *Connection) handleLogin(pktID int, r *Reader) {
	if pktID == 0x00 { // Login Start
		username, _ := r.ReadString()
		// Ignoring UUID for now
		log.Printf("Player '%s' is logging in.", username)

		// Create a player object
		c.p = &player.Player{
			Conn:     c.conn,
			Username: username,
			// UUID will be generated/assigned here
		}

		// Login Success (0x02)
		c.w.WritePacket(0x02,
			WriteString("00000000-0000-0000-0000-000000000000"), // Offline mode UUID
			WriteString(username),
			writeVarInt(0), // No properties
		)
		c.state = PlayState
		c.enterPlayState()
	}
}

func (c *Connection) enterPlayState() {
	// Send Join Game packet
	c.w.WritePacket(0x26, // Packet ID for Join Game
		WriteInt(0),                        // Player's Entity ID
		WriteBool(false),                   // Is hardcore
		WriteByte(1),                       // Gamemode (Creative)
		WriteByte(1),                       // Previous Gamemode
		writeVarInt(1),                     // World Count
		WriteString("minecraft:overworld"), // World Name
		WriteString("minecraft:overworld"), // Dimension Codec (placeholder)
		WriteLong(0),                       // Hashed seed
		writeVarInt(10),                    // Max players
		writeVarInt(8),                     // View distance
		writeVarInt(8),                     // Simulation distance
		WriteBool(false),                   // Reduced debug info
		WriteBool(true),                    // Enable respawn screen
		WriteBool(false),                   // Is debug
		WriteBool(false),                   // Is flat
	)

	// Keep-alive loop would start here
}
