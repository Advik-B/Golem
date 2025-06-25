package server

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/nbt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
)

const (
	ProtocolVersion = 767 // 1.21.5
	Version         = "Golem 1.21.5"

	// Serverbound (Client to Server)
	PacketIDClientInfo_C      = 0x00
	PacketIDPluginMessage_C   = 0x01 // Or Cookie Response
	PacketIDAckFinishConfig_C = 0x02
	PacketIDKeepAlive_C       = 0x03

	// Clientbound (Server to Client)
	PacketIDPluginMessage_S = 0x00
	PacketIDDisconnect_S    = 0x01
	PacketIDFinishConfig_S  = 0x02
	PacketIDKeepAlive_S     = 0x03
	PacketIDRegistryData_S  = 0x05
)

var ErrPingComplete = errors.New("ping sequence complete")

func (c *Connection) handleHandshake(p Packet) error {
	if p.ID != 0x00 {
		return fmt.Errorf("expected handshake packet (0x00), got 0x%X", p.ID)
	}
	data := p.Data
	_, bytesRead := readVarIntFromBytes(data) // Protocol Version
	data = data[bytesRead:]
	serverAddrLen, bytesRead := readVarIntFromBytes(data) // Server Address
	data = data[int(serverAddrLen)+bytesRead:]
	data = data[2:] // Skip Server Port
	nextState, _ := readVarIntFromBytes(data)

	addr := c.conn.RemoteAddr().String()
	if nextState == 1 {
		c.state = StateStatus
		Log.Debug("Transitioned to Status state", zap.String("remoteAddr", addr))
	} else if nextState == 2 {
		c.state = StateLogin
		Log.Debug("Transitioned to Login state", zap.String("remoteAddr", addr))
	} else {
		return fmt.Errorf("unknown next state: %d", nextState)
	}
	return nil
}

func (c *Connection) handleStatus(p Packet) error {
	switch p.ID {
	case 0x00:
		return c.sendStatusResponse()
	case 0x01:
		return c.handlePing(p)
	}
	return fmt.Errorf("unhandled packet in status state: 0x%X", p.ID)
}

func (c *Connection) handleLogin(p Packet) error {
	if p.ID != 0x00 { // Login Start
		return fmt.Errorf("expected login start packet (0x00), got 0x%X", p.ID)
	}

	playerName, _ := readStringFromBytes(p.Data)
	playerUUID := uuid.New()

	Log.Info("Player login attempt",
		zap.String("username", playerName),
		zap.Stringer("uuid", playerUUID),
		zap.String("remoteAddr", c.conn.RemoteAddr().String()),
	)

	// Add the player to the manager and associate it with THIS connection.
	c.server.PlayerManager.AddPlayer(c, playerName, playerUUID)

	// Send Login Success (0x02)
	packet := Packet{ID: 0x02}
	var buf bytes.Buffer
	buf.Write(writeUUID(playerUUID))
	buf.Write(writeString(playerName))
	buf.Write(writeVarInt(0)) // No properties
	packet.Data = buf.Bytes()

	if err := c.WritePacket(packet); err != nil {
		return fmt.Errorf("failed to send login success: %w", err)
	}

	// Transition the state of THIS connection.
	c.state = StateConfiguration
	return c.sendConfigurationPackets()
}

func (c *Connection) handleConfiguration(p Packet) error {
	remoteAddr := c.conn.RemoteAddr().String()
	switch p.ID {
	case PacketIDClientInfo_C: // 0x00: Client Information
		Log.Debug("Received Client Information packet", zap.String("remoteAddr", remoteAddr))
		return nil

	case PacketIDPluginMessage_C: // 0x01: Plugin Message / Cookie Response
		Log.Debug("Received Plugin Message/Cookie Response", zap.String("remoteAddr", remoteAddr))
		return nil

	case PacketIDAckFinishConfig_C: // 0x02: Acknowledge Finish Configuration (this is from the CLIENT)
		Log.Info("Client acknowledged configuration, transitioning to Play state",
			zap.String("remoteAddr", remoteAddr),
			zap.String("username", c.player.Username),
		)
		c.state = StatePlay
		if c.player == nil {
			return fmt.Errorf("cannot transition to play, no player associated with connection")
		}
		return c.sendJoinGamePackets(c.player.EntityID)

	case PacketIDKeepAlive_C: // 0x03: Keep-Alive (the one causing the crash)
		Log.Debug("Received and responding to configuration Keep-Alive", zap.String("remoteAddr", remoteAddr))
		// The client sent a Keep-Alive, we must respond with the same payload.
		// The response packet ID is also 0x03 in the configuration state.
		keepAliveResponse := Packet{
			ID:   PacketIDKeepAlive_S, // 0x03
			Data: p.Data,
		}
		return c.WritePacket(keepAliveResponse)

	default:
		Log.Warn("Skipping unknown configuration packet", zap.Int32("id", p.ID))
		Log.Debug("The packet in question: ", zap.Binary("data", p.Data))
		return nil
	}
}

// sendConfigurationPackets sends the server's initial configuration packets.
func (c *Connection) sendConfigurationPackets() error {
	// For now, we only need to send the "Finish Configuration" packet (0x02)
	// Other packets like Feature Flags and Registry Data would go here.
	finishConfigPacket := Packet{
		ID:   0x02,
		Data: []byte{}, // This packet has no data.
	}
	if err := c.WritePacket(finishConfigPacket); err != nil {
		return err
	}
	Log.Debug("Server sent Finish Configuration, waiting for client ack", zap.String("remoteAddr", c.conn.RemoteAddr().String()))
	return nil
}

func (c *Connection) handlePlay(p Packet) error {
	return nil
}

func (c *Connection) sendJoinGamePackets(entityID int32) error {
	var loginPacket bytes.Buffer

	loginPacket.Write(writeInt32(entityID))               // Player Entity ID
	loginPacket.Write(writeBool(false))                   // Is Hardcore
	loginPacket.Write(writeByte(1))                       // Gamemode: Creative
	loginPacket.Write(writeByte(255))                     // Previous Gamemode (255 = none)
	loginPacket.Write(writeVarInt(1))                     // World Count
	loginPacket.Write(writeString("minecraft:overworld")) // World Name (must match dimension key)

	// Dynamically build dimension codec
	dimensionCodec, err := buildDimensionCodec()
	if err != nil {
		return fmt.Errorf("could not build dimension codec: %w", err)
	}
	loginPacket.Write(dimensionCodec)

	loginPacket.Write(writeString("minecraft:overworld")) // Dimension type
	loginPacket.Write(writeLong(0))                       // Hashed seed
	loginPacket.Write(writeVarInt(8))                     // Max players
	loginPacket.Write(writeVarInt(10))                    // View distance
	loginPacket.Write(writeVarInt(10))                    // Simulation distance
	loginPacket.Write(writeBool(false))                   // Reduced debug info
	loginPacket.Write(writeBool(true))                    // Enable respawn screen
	loginPacket.Write(writeBool(false))                   // Is debug
	loginPacket.Write(writeBool(true))                    // Is flat
	loginPacket.Write(writeBool(false))                   // Has death screen
	loginPacket.Write(writeVarInt(0))                     // Portal cooldown (0 = instant)

	if err := c.WritePacket(Packet{ID: 0x29, Data: loginPacket.Bytes()}); err != nil {
		return err
	}

	// Set default spawn location (Position & Look)
	var spawnPosPacket bytes.Buffer
	spawnPosPacket.Write(writeBlockPos(0, 80, 0))
	spawnPosPacket.Write(writeFloat(0.0)) // angle
	if err := c.WritePacket(Packet{ID: 0x51, Data: spawnPosPacket.Bytes()}); err != nil {
		return err
	}

	// Set player abilities (Creative, Flying)
	var abilitiesPacket bytes.Buffer
	abilitiesPacket.Write(writeByte(0x0D))  // Flags
	abilitiesPacket.Write(writeFloat(0.05)) // Flying speed
	abilitiesPacket.Write(writeFloat(0.1))  // FOV modifier
	if err := c.WritePacket(Packet{ID: 0x36, Data: abilitiesPacket.Bytes()}); err != nil {
		return err
	}

	// Sync player position
	var posPacket bytes.Buffer
	posPacket.Write(writeDouble(0))
	posPacket.Write(writeDouble(80))
	posPacket.Write(writeDouble(0))
	posPacket.Write(writeFloat(0))  // Yaw
	posPacket.Write(writeFloat(0))  // Pitch
	posPacket.Write(writeByte(0))   // Flags (0 = absolute)
	posPacket.Write(writeVarInt(1)) // Teleport ID
	if err := c.WritePacket(Packet{ID: 0x3E, Data: posPacket.Bytes()}); err != nil {
		return err
	}

	// Send empty chunks (optional but safe)
	for x := -2; x <= 2; x++ {
		for z := -2; z <= 2; z++ {
			if err := c.WritePacket(generateEmptyChunkPacket(int32(x), int32(z))); err != nil {
				return err
			}
		}
	}

	// Set chunk center
	var centerPacket bytes.Buffer
	centerPacket.Write(writeVarInt(0))
	centerPacket.Write(writeVarInt(0))
	if err := c.WritePacket(Packet{ID: 0x50, Data: centerPacket.Bytes()}); err != nil {
		return err
	}

	Log.Info("Player has successfully joined the world",
		zap.String("username", c.player.Username),
		zap.Stringer("uuid", c.player.UUID),
	)
	return nil
}

func (c *Connection) sendStatusResponse() error {
	status := map[string]interface{}{
		"version": map[string]interface{}{
			"name":     Version,
			"protocol": ProtocolVersion,
		},
		"players": map[string]interface{}{
			"max":    100,
			"online": 0,
		},
		"description": map[string]interface{}{
			"text": "A Golem Server (Written in Go!)",
		},
	}
	jsonData, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("could not marshal status json: %w", err)
	}
	jsonStringBytes := writeString(string(jsonData))
	packet := Packet{
		ID:   0x00,
		Data: jsonStringBytes,
	}
	return c.WritePacket(packet)
}

// in internal/server/protocol.go
func (c *Connection) handlePing(p Packet) error {
	if len(p.Data) != 8 {
		return fmt.Errorf("invalid ping packet payload size")
	}
	responsePacket := Packet{
		ID:   0x01, // Pong
		Data: p.Data,
	}
	err := c.WritePacket(responsePacket)
	if err != nil {
		return err // If we can't write, it's a real error
	}

	Log.Debug("Responded to ping", zap.String("remoteAddr", c.conn.RemoteAddr().String()))

	// Signal a clean shutdown.
	return ErrPingComplete
}

func generateEmptyChunkPacket(x, z int32) Packet {
	var pktData bytes.Buffer
	pktData.Write(writeInt32(x))
	pktData.Write(writeInt32(z))

	heightmaps, _ := nbt.Marshal(map[string]interface{}{
		"MOTION_BLOCKING": make([]int64, 37),
	})
	pktData.Write(heightmaps)

	var dataBuf bytes.Buffer
	for i := 0; i < 24; i++ {
		dataBuf.Write(writeShort(0))
		dataBuf.WriteByte(0x00)
		dataBuf.Write(writeVarInt(0))
		dataBuf.Write(writeVarInt(0))
		dataBuf.WriteByte(0x00)
		dataBuf.Write(writeVarInt(131))
		dataBuf.Write(writeVarInt(0))
	}
	pktData.Write(writeVarInt(int32(dataBuf.Len())))
	pktData.Write(dataBuf.Bytes())
	pktData.Write(writeVarInt(0))
	pktData.Write(writeLong(0))
	pktData.Write(writeLong(0))
	pktData.Write(writeLong(0))
	pktData.Write(writeLong(0))
	pktData.Write(writeVarInt(0))
	pktData.Write(writeVarInt(0))

	return Packet{ID: 0x25, Data: pktData.Bytes()}
}

func buildDimensionCodec() ([]byte, error) {
	codec := map[string]interface{}{
		"minecraft:dimension_type": map[string]interface{}{
			"type": "minecraft:dimension_type",
			"value": []map[string]interface{}{
				{
					"name": "minecraft:overworld",
					"id":   int32(0),
					"element": map[string]interface{}{
						"piglin_safe":                     byte(0),
						"natural":                         byte(1),
						"ambient_light":                   float32(0.0),
						"infiniburn":                      "minecraft:infiniburn_overworld",
						"respawn_anchor_works":            byte(0),
						"has_skylight":                    byte(1),
						"bed_works":                       byte(1),
						"effects":                         "minecraft:overworld",
						"has_raids":                       byte(1),
						"min_y":                           int32(-64),
						"height":                          int32(384),
						"logical_height":                  int32(384),
						"coordinate_translation_scale":    float64(1.0),
						"ultrawarm":                       byte(0),
						"has_ceiling":                     byte(0),
						"monster_spawn_light_level":       int32(0),
						"monster_spawn_block_light_limit": int32(0),
					},
				},
			},
		},
		"minecraft:worldgen/biome": map[string]interface{}{
			"type": "minecraft:worldgen/biome",
			"value": []map[string]interface{}{
				{
					"name": "minecraft:plains",
					"id":   int32(1),
					"element": map[string]interface{}{
						"precipitation": "rain",
						"temperature":   float32(0.8),
						"downfall":      float32(0.4),
						"effects": map[string]interface{}{
							"sky_color":       int32(7907327),
							"water_fog_color": int32(329011),
							"fog_color":       int32(12638463),
							"water_color":     int32(4159204),
						},
					},
				},
			},
		},
	}
	return nbt.Marshal(codec)
}

// ----- VarInt Reading and Writing -----
func readVarInt(r io.Reader) (int32, error) {
	var numRead int
	var result int32
	var read byte
	for {
		b := make([]byte, 1)
		_, err := r.Read(b)
		if err != nil {
			return 0, err
		}
		read = b[0]
		value := int32(read & 0b01111111)
		result |= value << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, io.ErrUnexpectedEOF
		}
		if (read & 0b10000000) == 0 {
			break
		}
	}
	return result, nil
}

func readVarIntFromBytes(b []byte) (int32, int) {
	var numRead int
	var result int32
	var read byte
	for {
		if numRead >= len(b) {
			return 0, 0
		}
		read = b[numRead]
		value := int32(read & 0b01111111)
		result |= value << (7 * numRead)
		numRead++
		if numRead > 5 {
			return 0, 0
		}
		if (read & 0b10000000) == 0 {
			break
		}
	}
	return result, numRead
}

func writeVarInt(value int32) []byte {
	var buf []byte
	uValue := uint32(value)
	for {
		b := byte(uValue & 0b01111111)
		uValue >>= 7
		if uValue != 0 {
			b |= 0b10000000
		}
		buf = append(buf, b)
		if uValue == 0 {
			break
		}
	}
	return buf
}

// ----- Other Data Type Helpers -----

func readStringFromBytes(b []byte) (string, int) {
	strLen, bytesRead := readVarIntFromBytes(b)
	if bytesRead == 0 {
		return "", 0
	}
	end := bytesRead + int(strLen)
	if end > len(b) {
		return "", 0
	}
	return string(b[bytesRead:end]), end
}

func writeString(value string) []byte {
	strBytes := []byte(value)
	lenBytes := writeVarInt(int32(len(strBytes)))
	return append(lenBytes, strBytes...)
}

func writeUUID(value uuid.UUID) []byte {
	return value[:]
}

func writeInt(value int) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return buf
}

func writeInt32(value int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))
	return buf
}

func writeLong(value int64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(value))
	return buf
}

func writeBool(value bool) []byte {
	if value {
		return []byte{0x01}
	}
	return []byte{0x00}
}

func writeByte(value byte) []byte {
	return []byte{value}
}

func writeFloat(value float32) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, value)
	return buf.Bytes()
}

func writeDouble(value float64) []byte {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, value)
	return buf.Bytes()
}

func writeShort(value int16) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(value))
	return buf
}

func writeBlockPos(x, y, z int) []byte {
	val := (int64(x&0x3FFFFFF) << 38) | (int64(z&0x3FFFFFF) << 12) | int64(y&0xFFF)
	return writeLong(val)
}
