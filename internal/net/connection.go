package net

import (
	"bytes"
	"errors"
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
	conn          net.Conn
	state         int
	r             *Reader
	w             *Writer
	p             *player.Player
	addPlayerFunc func(*player.Player)
}

func NewConnection(conn net.Conn, addPlayerFunc func(*player.Player)) *Connection {
	return &Connection{
		conn:          conn,
		state:         HandshakeState,
		r:             NewReader(conn),
		w:             NewWriter(conn),
		addPlayerFunc: addPlayerFunc,
	}
}

func (c *Connection) Handle() (*player.Player, error) {
	for c.state != PlayState {
		pktLen, err := c.r.ReadVarInt()
		if err != nil {
			return nil, err
		}

		data := make([]byte, pktLen)
		if _, err := c.r.Read(data); err != nil {
			return nil, err
		}

		packetReader := NewReader(bytes.NewReader(data))
		pktID, _ := packetReader.ReadVarInt()

		switch c.state {
		case HandshakeState:
			c.handleHandshake(packetReader)
		case StatusState:
			c.handleStatus(pktID, packetReader)
			return nil, nil
		case LoginState:
			if err := c.handleLogin(pktID, packetReader); err != nil {
				return nil, err
			}
		}
	}
	return c.p, nil
}

func (c *Connection) HandlePlay() {
	for {
		pktLen, err := c.r.ReadVarInt()
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading packet length in play state: %v", err)
			}
			return
		}

		data := make([]byte, pktLen)
		if _, err := c.r.Read(data); err != nil {
			return
		}

		packetReader := NewReader(bytes.NewReader(data))
		pktID, _ := packetReader.ReadVarInt()

		if pktID == 0x12 { // ServerboundKeepAlivePacket ID
			id, _ := packetReader.ReadLong()
			if id != c.p.LastKeepAliveID {
				log.Printf("Invalid keep-alive ID from %s. Got %d, expected %d.", c.p.Username, id, c.p.LastKeepAliveID)
			}
		}
	}
}

func (c *Connection) handleHandshake(r *Reader) {
	_, _ = r.ReadVarInt()
	_, _ = r.ReadString()
	_, _ = r.r.ReadByte()
	_, _ = r.r.ReadByte()
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

func (c *Connection) handleLogin(pktID int, r *Reader) error {
	if pktID == 0x00 { // Login Start
		username, _ := r.ReadString()
		log.Printf("Player '%s' is logging in.", username)

		c.p = &player.Player{
			Conn:     c.conn,
			Username: username,
			State:    LoginState,
		}

		c.w.WritePacket(0x02,
			WriteString("00000000-0000-0000-0000-000000000000"),
			WriteString(username),
			WriteVarInt(0),
		)

		c.state = PlayState
		c.p.State = PlayState
		c.enterPlayState()
		c.addPlayerFunc(c.p)
		return nil
	}
	return errors.New("unexpected packet in login state")
}

func (c *Connection) enterPlayState() {
	// Join Game (ClientboundLoginPacket)
	c.w.WritePacket(0x26,
		WriteInt(1),
		WriteBool(false),
		WriteByte(1),
		WriteByte(0xFF),
		WriteVarInt(1),
		WriteString("minecraft:overworld"),
		WriteString("minecraft:overworld"),
		WriteLong(0),
		WriteVarInt(20),
		WriteVarInt(10),
		WriteVarInt(10),
		WriteBool(false),
		WriteBool(true),
		WriteBool(false),
		WriteBool(true),
		WriteBool(false),
	)

	// Brand (ClientboundCustomPayloadPacket)
	brandPayload := WriteString("Golem")
	c.w.WritePacket(0x18, // Packet ID for Custom Payload in Play state
		WriteString("minecraft:brand"),
		brandPayload,
	)

	// Difficulty (ClientboundChangeDifficultyPacket)
	c.w.WritePacket(0x0D,
		WriteByte(2),
		WriteBool(true),
	)

	// Player Abilities (ClientboundPlayerAbilitiesPacket)
	c.w.WritePacket(0x32,
		WriteByte(0x06), // can fly, is flying
		WriteFloat32(0.05),
		WriteFloat32(0.1),
	)

	// Set Held Item (ClientboundSetHeldItemPacket)
	c.w.WritePacket(0x4a, WriteByte(0))

	// Synchronize Player Position (ClientboundPlayerPositionPacket)
	c.w.WritePacket(0x38,
		WriteDouble(0.0),
		WriteDouble(64.0),
		WriteDouble(0.0),
		WriteFloat32(0.0),
		WriteFloat32(0.0),
		WriteByte(0),
		WriteVarInt(1),
	)
}
