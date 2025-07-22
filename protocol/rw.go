// File: protocol/rw.go
package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

const (
	maxVarIntLen  = 5
	maxVarLongLen = 10
	segmentBits   = 0x7F
	continueBit   = 0x80
)

// --- Primitive Readers ---

func ReadBool(r io.Reader) (bool, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return false, err
	}
	return b[0] != 0, nil
}

func ReadUint8(r io.Reader) (uint8, error) {
	var b [1]byte
	if _, err := io.ReadFull(r, b[:]); err != nil {
		return 0, err
	}
	return b[0], nil
}

func ReadInt8(r io.Reader) (int8, error) {
	val, err := ReadUint8(r)
	return int8(val), err
}

func ReadUint16(r io.Reader) (uint16, error) {
	var val uint16
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadInt16(r io.Reader) (int16, error) {
	var val int16
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadUint32(r io.Reader) (uint32, error) { // Added this function
	var val uint32
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadInt32(r io.Reader) (int32, error) {
	var val int32
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadInt64(r io.Reader) (int64, error) {
	var val int64
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadFloat32(r io.Reader) (float32, error) {
	var val float32
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

func ReadFloat64(r io.Reader) (float64, error) {
	var val float64
	err := binary.Read(r, binary.BigEndian, &val)
	return val, err
}

// --- Primitive Writers ---

func WriteBool(w io.Writer, v bool) error {
	var b byte = 0
	if v {
		b = 1
	}
	_, err := w.Write([]byte{b})
	return err
}

func WriteUint8(w io.Writer, v uint8) error {
	_, err := w.Write([]byte{v})
	return err
}

func WriteInt8(w io.Writer, v int8) error {
	return WriteUint8(w, uint8(v))
}

func WriteUint16(w io.Writer, v uint16) error {
	return binary.Write(w, binary.BigEndian, v)
}

func WriteInt16(w io.Writer, v int16) error {
	return binary.Write(w, binary.BigEndian, v)
}

func WriteUint32(w io.Writer, v uint32) error { // Added this function
	return binary.Write(w, binary.BigEndian, v)
}

func WriteInt32(w io.Writer, v int32) error {
	return binary.Write(w, binary.BigEndian, v)
}

func WriteInt64(w io.Writer, v int64) error {
	return binary.Write(w, binary.BigEndian, v)
}

func WriteFloat32(w io.Writer, v float32) error {
	return binary.Write(w, binary.BigEndian, v)
}

func WriteFloat64(w io.Writer, v float64) error {
	return binary.Write(w, binary.BigEndian, v)
}

// --- Minecraft-specific Types ---

func ReadVarInt(r io.Reader) (int32, error) {
	var value int32
	var position int
	var currentByte byte
	var b [1]byte

	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		currentByte = b[0]
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

func WriteVarInt(w io.Writer, value int32) error {
	v := uint32(value)
	for {
		if (v & ^uint32(segmentBits)) == 0 {
			return WriteUint8(w, byte(v))
		}
		if err := WriteUint8(w, byte((v&segmentBits)|continueBit)); err != nil {
			return err
		}
		v >>= 7
	}
}

func ReadVarLong(r io.Reader) (int64, error) {
	var value int64
	var position int
	var currentByte byte
	var b [1]byte

	for {
		if _, err := io.ReadFull(r, b[:]); err != nil {
			return 0, err
		}
		currentByte = b[0]
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

func WriteVarLong(w io.Writer, value int64) error {
	v := uint64(value)
	for {
		if (v & ^uint64(segmentBits)) == 0 {
			return WriteUint8(w, byte(v))
		}
		if err := WriteUint8(w, byte((v&segmentBits)|continueBit)); err != nil {
			return err
		}
		v >>= 7
	}
}

func ReadString(r io.Reader) (string, error) {
	length, err := ReadVarInt(r)
	if err != nil {
		return "", err
	}
	if length < 0 {
		return "", fmt.Errorf("string length is negative: %d", length)
	}

	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}

	return string(buf), nil
}

func WriteString(w io.Writer, value string) error {
	if err := WriteVarInt(w, int32(len(value))); err != nil {
		return err
	}
	_, err := w.Write([]byte(value))
	return err
}

func ReadUUID(r io.Reader) (uuid.UUID, error) {
	var u uuid.UUID
	if _, err := io.ReadFull(r, u[:]); err != nil {
		return uuid.Nil, err
	}
	return u, nil
}

func WriteUUID(w io.Writer, u uuid.UUID) error {
	_, err := w.Write(u[:])
	return err
}

func ReadNBT(r io.Reader) (nbt.Tag, error) {
	// Minecraft can send a single 0x00 byte to indicate no NBT.
	// We peek at the first byte to check for this case.
	// For this, we need a reader that supports peeking.
	br, ok := r.(io.ByteReader)
	if !ok {
		// Fallback for readers that don't support ByteReader.
		// This is less efficient as it involves a temporary buffer.
		var firstByte [1]byte
		if _, err := io.ReadFull(r, firstByte[:]); err != nil {
			return nil, err // Could be EOF for optional NBT
		}
		if firstByte[0] == nbt.TagEnd {
			return nil, nil // No NBT
		}
		// Re-join the byte to the stream for the NBT parser.
		r = io.MultiReader(strings.NewReader(string(firstByte[:])), r)
	} else {
		firstByte, err := br.ReadByte()
		if err != nil {
			return nil, err
		}
		if firstByte == nbt.TagEnd {
			return nil, nil
		}
		r = io.MultiReader(strings.NewReader(string(firstByte)), r.(io.Reader))
	}

	namedTag, err := nbt.Read(r)
	if err != nil {
		return nil, err
	}
	return namedTag.Tag, nil
}

func WriteNBT(w io.Writer, tag nbt.Tag) error {
	if tag == nil {
		return WriteUint8(w, nbt.TagEnd)
	}
	// NBT is written as a named tag, but root tag often has an empty name.
	return nbt.Write(w, nbt.NamedTag{Name: "", Tag: tag})
}

func ReadPosition(r io.Reader) (x, y, z int, err error) {
	val, err := ReadInt64(r)
	if err != nil {
		return 0, 0, 0, err
	}
	x = int(val >> 38)
	y = int(val & 0xFFF)
	z = int(val << 26 >> 38)
	return
}

func WritePosition(w io.Writer, x, y, z int) error {
	val := (int64(x&0x3FFFFFF) << 38) | (int64(z&0x3FFFFFF) << 12) | int64(y&0xFFF)
	return WriteInt64(w, val)
}
