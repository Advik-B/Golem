package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"net"
)

const (
	HandshakeState = 0
	StatusState    = 1
	LoginState     = 2
	PlayState      = 3
	ProtocolVer    = 765 // Minecraft 1.21.5
)

func main() {
	ln, err := net.Listen("tcp", ":25565")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Listening on :25565")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r := bufio.NewReader(conn)

	state := HandshakeState
	var username string

	for {
		pktLen, _ := readVarInt(r)
		data := make([]byte, pktLen)
		if _, err := io.ReadFull(r, data); err != nil {
			return
		}
		pkt := bufio.NewReader(io.NopCloser(bufio.NewReaderSize(bytesReader(data), len(data))))
		pktID, _ := readVarInt(pkt)

		switch state {
		case HandshakeState:
			_, _ = readVarInt(pkt)         // protocol version
			_, _ = readString(pkt)         // server address
			_, _ = binary.ReadUvarint(pkt) // port
			nextState, _ := readVarInt(pkt)
			state = nextState

		case StatusState:
			// status request (0x00)
			if pktID == 0x00 {
				resp := `{"version":{"name":"1.21.5","protocol":765},"players":{"max":1,"online":0},"description":{"text":"Void"}}`
				writePacket(conn, 0x00, writeString(resp))
			}
		case LoginState:
			if pktID == 0x00 {
				username, _ = readString(pkt)
				// Login success (0x02)
				writePacket(conn, 0x02,
					writeString("00000000-0000-0000-0000-000000000000"),
					writeString(username),
				)
				state = PlayState

				// Send Join Game
				writePacket(conn, 0x26,
					writeInt(1),                // Entity ID
					writeByte(0), writeByte(0), // Game mode + previous
					writeVarInt(1), writeString("minecraft:overworld"), // World count + name
					writeString("minecraft:overworld"),              // dimension codec
					writeLong(0),                                    // Hashed seed
					writeVarInt(10), writeVarInt(0), writeVarInt(0), // max players, view/sim dist
					writeBool(true), writeBool(true), writeBool(false), writeBool(false),
				)

				// Send empty chunk (0, 0)
				writePacket(conn, 0x22,
					writeInt(0), writeInt(0), // chunk XZ
					writeBool(true), writeVarInt(0), writeVarInt(1),
					[]byte{0x00},   // chunk data
					writeVarInt(0), // block entities
					writeBool(true),
					writeVarInt(0), writeVarInt(0),
					[]byte{}, []byte{}, // light arrays
				)
			}
		}
	}
}

func writePacket(w io.Writer, id int, payload ...[]byte) {
	p := writeVarInt(id)
	for _, part := range payload {
		p = append(p, part...)
	}
	final := writeVarInt(len(p))
	final = append(final, p...)
	w.Write(final)
}

func readVarInt(r *bufio.Reader) (int, error) {
	var num int
	var shift uint
	for {
		b, err := r.ReadByte()
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

func writeString(s string) []byte {
	out := writeVarInt(len(s))
	out = append(out, []byte(s)...)
	return out
}

func readString(r *bufio.Reader) (string, error) {
	length, err := readVarInt(r)
	if err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func writeInt(i int) []byte {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], uint32(i))
	return buf[:]
}

func writeLong(i int64) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(i))
	return buf[:]
}

func writeByte(b byte) []byte { return []byte{b} }
func writeBool(b bool) []byte {
	if b {
		return []byte{1}
	}
	return []byte{0}
}

func bytesReader(b []byte) *bufio.Reader {
	return bufio.NewReaderSize(io.NopCloser(io.LimitReader(io.MultiReader(bytes.NewReader(b)), int64(len(b)))), len(b))
}
