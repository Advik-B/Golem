package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/panjf2000/gnet/v2"
)

// VarIntFrameCodec implements gnet.ICodec for Minecraft's VarInt-prefixed packets.
type VarIntFrameCodec struct{}

// Encode prepends the VarInt length to the given buffer.
// The input `buf` should contain the complete packet payload (ID + Data).
func (vc *VarIntFrameCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	packetLen := int32(len(buf))
	varIntBuf := make([]byte, 5) // 5 bytes is the max size for a VarInt.
	n := binary.PutUvarint(varIntBuf, uint64(packetLen))
	return append(varIntBuf[:n], buf...), nil
}

// Decode reads one full packet frame from the connection's inbound buffer.
func (vc *VarIntFrameCodec) Decode(c gnet.Conn) ([]byte, error) {
	// Peek up to 5 bytes to read the VarInt without consuming data from the buffer.
	header, _ := c.Peek(5)
	if len(header) == 0 {
		return nil, nil // Not enough data to read header
	}

	packetLen, varIntLen := binary.Uvarint(header)
	if varIntLen <= 0 {
		return nil, fmt.Errorf("failed to read varint length, read bytes: %d", varIntLen)
	}

	if packetLen > 2097151 { // Max packet size in Minecraft
		return nil, fmt.Errorf("packet length %d exceeds max size", packetLen)
	}

	fullPacketLen := int(packetLen) + varIntLen
	if c.InboundBuffered() < fullPacketLen {
		return nil, nil // Not enough data for the full frame, wait for more.
	}

	// Discard the length prefix from the connection's inbound buffer.
	_, _ = c.Discard(varIntLen)
	// Read the exact length of the packet payload.
	return c.Next(int(packetLen))
}
