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
	// Create a temporary buffer for the VarInt prefix.
	// 5 bytes is the max size for a VarInt.
	varIntBuf := make([]byte, 5) // Correctly named buffer
	n := binary.PutUvarint(varIntBuf, uint64(packetLen))

	// Prepend the VarInt to the original buffer.
	return append(varIntBuf[:n], buf...), nil
}

// Decode reads a full packet frame from the connection's inbound buffer.
func (vc *VarIntFrameCodec) Decode(c gnet.Conn) ([]byte, error) {
	// The gnet-dev source shows that `Peek` is safe to use for reading the header.
	header, _ := c.Peek(5)
	if len(header) == 0 {
		return nil, nil // Not enough data to read header
	}

	// Decode the VarInt to find the packet length and the size of the length prefix itself.
	packetLen, varIntLen := binary.Uvarint(header)
	if varIntLen <= 0 {
		return nil, fmt.Errorf("failed to read varint length, read bytes: %d", varIntLen)
	}

	// Max packet size in Minecraft is 2MB + 1 byte
	if packetLen > 2097151 {
		return nil, fmt.Errorf("packet length %d exceeds max size", packetLen)
	}

	fullPacketLen := int(packetLen) + varIntLen
	if c.InboundBuffered() < fullPacketLen {
		return nil, nil // Not enough data, wait for more.
	}

	// Discard the length prefix from the connection's inbound buffer.
	_, _ = c.Discard(varIntLen)
	// Read the exact length of the packet payload.
	return c.Next(int(packetLen))
}
