package net

import (
	"bufio"
	"crypto/cipher"
	"encoding/binary"
	"io"
	"math"
)

type Reader struct {
	r         *bufio.Reader
	decStream cipher.Stream
}

//func (r *Reader) EnableEncryption(secret []byte) {
//	block, _ := aes.NewCipher(secret)
//	r.decStream = cipher.NewCFB8Decrypter(block, secret)
//}

func (r *Reader) Read(p []byte) (int, error) {
	n, err := io.ReadFull(r.r, p)
	if err == nil && r.decStream != nil {
		r.decStream.XORKeyStream(p[:n], p[:n])
	}
	return n, err
}

type Writer struct {
	w         io.Writer
	encStream cipher.Stream
}

//func (w *Writer) EnableEncryption(block cipher.Block, secret []byte) {
//	w.encStream = cipher.NewCFB8Encrypter(block, secret)
//}

func (w *Writer) Write(p []byte) (int, error) {
	if w.encStream != nil {
		buf := make([]byte, len(p))
		w.encStream.XORKeyStream(buf, p)
		return w.w.Write(buf)
	}
	return w.w.Write(p)
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) ReadByte() (byte, error) {
	return r.r.ReadByte()
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
		if shift >= 35 {
			return 0, io.ErrUnexpectedEOF
		}
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

func (r *Reader) ReadLong() (int64, error) {
	var val int64
	err := binary.Read(r.r, binary.BigEndian, &val)
	return val, err
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

func (w *Writer) WritePacket(id int, payload ...[]byte) {
	p := WriteVarInt(id)
	for _, part := range payload {
		p = append(p, part...)
	}
	final := WriteVarInt(len(p))
	final = append(final, p...)
	w.w.Write(final)
}

func WriteVarInt(n int) []byte {
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
	out := WriteVarInt(len(s))
	out = append(out, []byte(s)...)
	return out
}

func WriteInt(i int32) []byte {
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

func WriteDouble(d float64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], math.Float64bits(d))
	return buf[:]
}

func WriteFloat32(f float32) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], math.Float32bits(f))
	return buf[:]
}
