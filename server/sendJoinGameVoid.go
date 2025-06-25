package server

/*
func sendJoinGameVoid(c *Connection, playerEntityID int32) error {
	// === 1. Join Game Packet ===
	join := NewPacket(0x29) // Join Game

	join.WriteInt(playerEntityID) // Player Entity ID
	join.WriteBool(false)         // Is Hardcore
	join.WriteByte(1)             // Gamemode (Creative)
	join.WriteByte(1)             // Previous Gamemode
	join.WriteVarInt(1)           // World count
	join.WriteString("void")      // World name (must match dimension below)

	join.WriteNbt(minimalDimensionCodec()) // Dimension codec (hardcoded NBT)

	join.WriteString("minecraft:overworld") // Dimension type
	join.WriteString("void")                // World name (arbitrary)
	join.WriteLong(0)                       // Hashed seed
	join.WriteVarInt(0)                     // Max players (unused)
	join.WriteVarInt(8)                     // View distance
	join.WriteVarInt(8)                     // Simulation distance
	join.WriteBool(false)                   // Reduced debug info
	join.WriteBool(false)                   // Enable respawn screen
	join.WriteBool(false)                   // Is Debug
	join.WriteBool(true)                    // Is Flat

	if err := c.WritePacket(join); err != nil {
		return err
	}

	// === 2. Player Position and Look ===
	pos := protocol.NewPacket(0x43) // Player Position and Look
	pos.WriteDouble(0)              // X
	pos.WriteDouble(64)             // Y (spawn high enough)
	pos.WriteDouble(0)              // Z
	pos.WriteFloat(0)               // Yaw
	pos.WriteFloat(0)               // Pitch
	pos.WriteByte(0)                // Relative flags
	pos.WriteVarInt(0)              // Teleport ID
	pos.WriteBool(false)            // Dismount vehicle

	return c.WritePacket(pos)
}
*/
