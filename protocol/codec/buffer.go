package codec

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/google/uuid"
)

// Constants for VarInt/VarLong encoding.
const (
	segmentBits = 0x7F
	continueBit = 0x80
)

// PacketBuffer wraps a bytes.Buffer to provide methods for reading and writing
// Minecraft protocol data types.
type PacketBuffer struct {
	*bytes.Buffer
}

// NewPacketBuffer creates a new PacketBuffer wrapping the given byte slice.
func NewPacketBuffer(buf []byte) *PacketBuffer {
	return &PacketBuffer{bytes.NewBuffer(buf)}
}

// Read and Write methods to satisfy io.ReadWriter
func (b *PacketBuffer) Read(p []byte) (n int, err error)  { return b.Buffer.Read(p) }
func (b *PacketBuffer) Write(p []byte) (n int, err error) { return b.Buffer.Write(p) }

// --- Primitive Readers/Writers ---

func (b *PacketBuffer) ReadBool() (bool, error) { val, err := b.ReadByte(); return val != 0, err }
func (b *PacketBuffer) WriteBool(v bool) error {
	var val byte = 0
	if v {
		val = 1
	}
	return b.WriteByte(val)
}

func (b *PacketBuffer) ReadByte() (byte, error) { return b.Buffer.ReadByte() }
func (b *PacketBuffer) WriteByte(v byte) error  { return b.Buffer.WriteByte(v) }

func (b *PacketBuffer) ReadShort() (int16, error) {
	var v int16
	err := binary.Read(b, binary.BigEndian, &v)
	return v, err
}
func (b *PacketBuffer) WriteShort(v int16) error { return binary.Write(b, binary.BigEndian, v) }

func (b *PacketBuffer) ReadInt() (int32, error) {
	var v int32
	err := binary.Read(b, binary.BigEndian, &v)
	return v, err
}
func (b *PacketBuffer) WriteInt(v int32) error { return binary.Write(b, binary.BigEndian, v) }

func (b *PacketBuffer) ReadLong() (int64, error) {
	var v int64
	err := binary.Read(b, binary.BigEndian, &v)
	return v, err
}
func (b *PacketBuffer) WriteLong(v int64) error { return binary.Write(b, binary.BigEndian, v) }

func (b *PacketBuffer) ReadFloat() (float32, error) {
	var v uint32
	err := binary.Read(b, binary.BigEndian, &v)
	return math.Float32frombits(v), err
}
func (b *PacketBuffer) WriteFloat(v float32) error {
	return binary.Write(b, binary.BigEndian, math.Float32bits(v))
}

func (b *PacketBuffer) ReadDouble() (float64, error) {
	var v uint64
	err := binary.Read(b, binary.BigEndian, &v)
	return math.Float64frombits(v), err
}
func (b *PacketBuffer) WriteDouble(v float64) error {
	return binary.Write(b, binary.BigEndian, math.Float64bits(v))
}

// --- VarInt / VarLong ---

func (b *PacketBuffer) ReadVarInt() (int32, error) {
	var value int32
	var position int
	for {
		currentByte, err := b.ReadByte()
		if err != nil {
			return 0, err
		}
		value |= int32(currentByte&segmentBits) << position
		if (currentByte & continueBit) == 0 {
			break
		}
		position += 7
		if position >= 32 {
			return 0, fmt.Errorf("VarInt is too big")
		}
	}
	return value, nil
}

func (b *PacketBuffer) WriteVarInt(value int32) error {
	uvalue := uint32(value)
	for {
		if (uvalue & ^uint32(segmentBits)) == 0 {
			return b.WriteByte(byte(uvalue))
		}
		if err := b.WriteByte(byte(uvalue&segmentBits | continueBit)); err != nil {
			return err
		}
		uvalue >>= 7
	}
}

func (b *PacketBuffer) ReadVarLong() (int64, error) {
	var value int64
	var position int
	for {
		currentByte, err := b.ReadByte()
		if err != nil {
			return 0, err
		}
		value |= int64(currentByte&segmentBits) << position
		if (currentByte & continueBit) == 0 {
			break
		}
		position += 7
		if position >= 64 {
			return 0, fmt.Errorf("VarLong is too big")
		}
	}
	return value, nil
}

func (b *PacketBuffer) WriteVarLong(value int64) error {
	uvalue := uint64(value)
	for {
		if (uvalue & ^uint64(segmentBits)) == 0 {
			return b.WriteByte(byte(uvalue))
		}
		if err := b.WriteByte(byte(uvalue&segmentBits | continueBit)); err != nil {
			return err
		}
		uvalue >>= 7
	}
}

// --- String ---

func (b *PacketBuffer) ReadString(maxLength int) (string, error) {
	length, err := b.ReadVarInt()
	if err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}
	if length < 0 {
		return "", fmt.Errorf("string length is negative: %d", length)
	}
	if length > int32(maxLength) {
		return "", fmt.Errorf("string length %d exceeds max length %d", length, maxLength)
	}
	if b.Len() < int(length) {
		return "", fmt.Errorf("not enough bytes for string, expected %d, have %d", length, b.Len())
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(b, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func (b *PacketBuffer) WriteString(s string) error {
	if err := b.WriteVarInt(int32(len(s))); err != nil {
		return err
	}
	_, err := b.Buffer.WriteString(s)
	return err
}

// --- Other Complex Types ---

func (b *PacketBuffer) ReadUUID() (uuid.UUID, error) {
	var u uuid.UUID
	_, err := io.ReadFull(b, u[:])
	return u, err
}

func (b *PacketBuffer) WriteUUID(u uuid.UUID) error {
	_, err := b.Write(u[:])
	return err
}

func (b *PacketBuffer) ReadByteArray() ([]byte, error) {
	length, err := b.ReadVarInt()
	if err != nil {
		return nil, err
	}
	if length < 0 || length > int32(b.Len()) {
		return nil, fmt.Errorf("invalid byte array length: %d", length)
	}
	buf := make([]byte, length)
	_, err = io.ReadFull(b, buf)
	return buf, err
}

func (b *PacketBuffer) WriteByteArray(data []byte) error {
	if err := b.WriteVarInt(int32(len(data))); err != nil {
		return err
	}
	_, err := b.Write(data)
	return err
}
