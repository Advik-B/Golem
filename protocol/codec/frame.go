package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/panjf2000/gnet/v2"
)

// VarIntFrameCodec implements gnet.ICodec for Minecraft's VarInt-prefixed packets.
type VarIntFrameCodec struct{}

func (vc *VarIntFrameCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	packetLen := int32(len(buf))
	// Create a temporary buffer for the VarInt prefix.
	// 5 bytes is the max size for a VarInt.
	varIntBuf := make([]byte, 5)
	n := binary.PutUvarint(varIntBuf, uint64(packetLen))

	// Prepend the VarInt to the original buffer.
	return append(varIntBuf[:n], buf...), nil
}

func (vc *VarIntFrameCodec) Decode(c gnet.Conn) ([]byte, error) {
	// Peek the header to read the VarInt length without consuming the bytes from the buffer.
	header, _ := c.Peek(5)
	if len(header) == 0 {
		return nil, nil // Not enough data to read header
	}

	// Decode the VarInt to find the packet length and the size of the length prefix itself.
	packetLen, varIntLen := binary.Uvarint(header)
	if varIntLen <= 0 {
		return nil, fmt.Errorf("failed to read varint length, read bytes: %d", varIntLen)
	}

	// Check for Minecraft's max packet size.
	if packetLen > 2097151 {
		return nil, fmt.Errorf("packet length %d exceeds max size", packetLen)
	}

	fullPacketLen := int(packetLen) + varIntLen
	// Check if the full packet (prefix + payload) has arrived.
	if c.InboundBuffered() < fullPacketLen {
		return nil, nil // Not enough data, wait for more.
	}

	// Discard the length prefix from the connection's inbound buffer.
	c.Discard(varIntLen)
	// Read the exact length of the packet payload.
	payload, err := c.Next(int(packetLen))
	if err != nil {
		// This should theoretically not happen if the previous checks pass.
		return nil, fmt.Errorf("failed to read full packet payload: %w", err)
	}
	return payload, nil
}
