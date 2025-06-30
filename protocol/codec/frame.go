package codec

import (
	"encoding/binary"
	"fmt"
	"github.com/panjf2000/gnet/v2"
)

// VarIntFrameCodec implements gnet.ICodec for Minecraft's VarInt-prefixed packets.
type VarIntFrameCodec struct{}

func (vc *VarIntFrameCodec) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	// The PacketBuffer already writes the VarInt length, so here we just need to prepend it
	// to the packet payload (ID + data).
	packetLen := int32(len(buf))
	varIntLenBuf := make([]byte, 5) // Max 5 bytes for a VarInt
	n := binary.PutUvarint(varIntLenBuf, uint64(packetLen))

	return append(varIntLenBuf[:n], buf...), nil
}

func (vc *VarIntFrameCodec) Decode(c gnet.Conn) ([]byte, error) {
	// Peek the header to read the VarInt length without consuming the bytes.
	header, _ := c.Peek(5)
	if len(header) == 0 {
		return nil, nil
	}

	packetLen, varIntLen := binary.Uvarint(header)
	if varIntLen <= 0 {
		return nil, fmt.Errorf("failed to read varint length, read bytes: %d", varIntLen)
	}
	if packetLen > 2097151 { // Max packet size
		return nil, fmt.Errorf("packet length %d exceeds max size", packetLen)
	}

	// Check if the full packet has arrived.
	fullPacketLen := int(packetLen) + varIntLen
	if c.InboundBuffered() < fullPacketLen {
		return nil, nil // Not enough data, wait for more.
	}

	// Discard the VarInt length prefix from the buffer.
	c.Discard(varIntLen)
	// Read the full packet payload.
	return c.ReadN(int(packetLen))
}
