package server

import (
	"fmt"
	"io"
)

// Packet represents a generic Minecraft packet.
type Packet struct {
	ID   int32
	Data []byte
}

// ReadPacket reads a full packet from the connection.
func (c *Connection) ReadPacket() (Packet, error) {
	// Read the packet length (VarInt)
	length, err := readVarInt(c.conn)
	if err != nil {
		return Packet{}, fmt.Errorf("could not read packet length: %w", err)
	}

	if length <= 0 {
		return Packet{}, fmt.Errorf("invalid packet length: %d", length)
	}

	// Read the full packet data (ID + Data) into a buffer
	buffer := make([]byte, length)
	_, err = io.ReadFull(c.conn, buffer)
	if err != nil {
		return Packet{}, fmt.Errorf("could not read packet data: %w", err)
	}

	// The first part of the buffer is the packet ID (VarInt)
	packetID, bytesRead := readVarIntFromBytes(buffer)

	// The rest is the actual data
	data := buffer[bytesRead:]

	return Packet{ID: packetID, Data: data}, nil
}

// WritePacket serializes and sends a packet to the client.
func (c *Connection) WritePacket(p Packet) error {
	packetIDBytes := writeVarInt(p.ID)

	// Packet Length = length of Packet ID + length of Data
	packetLength := len(packetIDBytes) + len(p.Data)

	lengthBytes := writeVarInt(int32(packetLength))

	// Write the full packet to the connection
	_, err := c.conn.Write(append(lengthBytes, append(packetIDBytes, p.Data...)...))
	return err
}
