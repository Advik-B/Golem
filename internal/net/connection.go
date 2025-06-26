package net

import (
	"bytes"
	"errors"
	"github.com/Advik-B/Golem/internal/player"
	"github.com/Advik-B/Golem/nbt"
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
	// ClientboundLoginPacket (0x26)
	// We now construct this packet with the correct fields for 1.21.
	codecTag := GetDimensionCodec()
	var codecBuf bytes.Buffer
	// The dimension codec is sent as a standalone NBT tag.
	if err := nbt.Write(&codecBuf, nbt.NamedTag{Name: "", Tag: codecTag}); err != nil {
		panic("failed to write dimension codec to buffer: " + err.Error())
	}

	c.w.WritePacket(0x26,
		WriteInt(1),      // entityId
		WriteBool(false), // is hardcore
		// World Names (Array of Identifier)
		WriteVarInt(1), // One world
		WriteString("minecraft:overworld"),
		// End of World Names
		WriteVarInt(20),  // maxPlayers
		WriteVarInt(10),  // viewDistance
		WriteVarInt(10),  // simulationDistance
		WriteBool(false), // reducedDebugInfo
		WriteBool(true),  // enableRespawnScreen
		WriteBool(false), // isDebug
		WriteBool(true),  // isFlat

		// CommonPlayerSpawnInfo starts here
		codecBuf.Bytes(),                   // Dimension Codec NBT
		WriteString("minecraft:overworld"), // Dimension Name
		WriteLong(0),                       // Hashed seed
		WriteByte(1),                       // gameType (Creative)
		WriteByte(0xFF),                    // previousGameType (-1 for none)
		WriteBool(false),                   // isDebug (again, for spawn info)
		WriteBool(true),                    // isFlat (again, for spawn info)
		WriteBool(false),                   // No last death location
		WriteVarInt(0),                     // portalCooldown
	)

	// ClientboundCustomPayloadPacket (Brand) - ID is 0x19 in Play state
	brandPayload := WriteString("Golem")
	c.w.WritePacket(0x19,
		WriteString("minecraft:brand"),
		brandPayload,
	)

	// ClientboundChangeDifficultyPacket
	c.w.WritePacket(0x0E,
		WriteByte(1),    // Difficulty (easy)
		WriteBool(true), // Locked
	)

	// ClientboundPlayerAbilitiesPacket
	c.w.WritePacket(0x32,
		WriteByte(0x06),    // Flags (invulnerable, flying)
		WriteFloat32(0.05), // Flying Speed
		WriteFloat32(0.1),  // FOV Modifier
	)

	// ClientboundSetHeldItemPacket
	c.w.WritePacket(0x4A, WriteByte(0)) // slot 0

	// ClientboundPlayerPositionPacket
	c.w.WritePacket(0x3E, // This was 0x38, correct ID for 1.21 is 0x3E
		WriteDouble(0.0),   // X
		WriteDouble(100.0), // Y
		WriteDouble(0.0),   // Z
		WriteFloat32(0.0),  // Yaw
		WriteFloat32(0.0),  // Pitch
		WriteByte(0),       // flags
		WriteVarInt(1),     // teleport id
	)
}
