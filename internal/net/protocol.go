package net

import (
	"bufio"
	"encoding/binary"
	"io"
)

// Reader is a helper for reading Minecraft protocol data.
type Reader struct {
	r *bufio.Reader
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	return io.ReadFull(r.r, p)
}

func (r *Reader) ReadVarInt() (int, error) {
	var num int
	var shift uint
	for {
		b, err := r.r.ReadByte()
		if err != nil {
			return 0, err
		}
		num |= int(b&0x7F) << shift
		if (b & 0x80) == 0 {
			break
		}
		shift += 7
	}
	return num, nil
}

func (r *Reader) ReadString() (string, error) {
	length, err := r.ReadVarInt()
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r.r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// Writer is a helper for writing Minecraft protocol data.
type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

// WritePacket sends a packet with its ID and payload, handling length prefixing.
func (w *Writer) WritePacket(id int, payload ...[]byte) {
	p := writeVarInt(id)
	for _, part := range payload {
		p = append(p, part...)
	}
	final := writeVarInt(len(p))
	final = append(final, p...)
	w.w.Write(final)
}

// --- Internal helpers for writing specific types ---

func writeVarInt(n int) []byte {
	var out []byte
	for {
		b := byte(n & 0x7F)
		n >>= 7
		if n != 0 {
			b |= 0x80
		}
		out = append(out, b)
		if n == 0 {
			break
		}
	}
	return out
}

func WriteString(s string) []byte {
	out := writeVarInt(len(s))
	out = append(out, []byte(s)...)
	return out
}

func WriteInt(i int) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	return buf[:]
}

func WriteLong(i int64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func WriteByte(b byte) []byte { return []byte{b} }

func WriteBool(b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}
