package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

const (
	protocolVer = 770 // 1.21.5
	motd        = "§aHello Void!"
	maxPlayers  = 20
	serverPort  = 25565
)

// VarInt helpers
func readVarInt(r io.Reader) (int32, error) {
	var (
		num   int32
		shift uint
	)
	buf := make([]byte, 1)
	for {
		if _, err := r.Read(buf); err != nil {
			return 0, err
		}
		b := buf[0]
		num |= int32(b&0x7F) << shift
		if b&0x80 == 0 {
			return num, nil
		}
		shift += 7
	}
}

func writeVarInt(w io.Writer, val int32) error {
	for {
		if val&^0x7F == 0 {
			_, err := w.Write([]byte{byte(val)})
			return err
		}
		_, err := w.Write([]byte{byte(val&0x7F | 0x80)})
		if err != nil {
			return err
		}
		val >>= 7
	}
}

// Packet helpers
func sendPacket(w net.Conn, id int32, payload []byte) {
	var buf bytes.Buffer
	writeVarInt(&buf, id)
	buf.Write(payload)

	var frame bytes.Buffer
	writeVarInt(&frame, int32(buf.Len()))
	frame.Write(buf.Bytes())
	w.Write(frame.Bytes())
}

func writeString(w *bytes.Buffer, s string) {
	writeVarInt(w, int32(len(s)))
	w.WriteString(s)
}

// Registry/NBT building
func buildDimensionCodec() []byte {
	root := nbt.NewCompoundTag()

	dimType := nbt.NewCompoundTag()
	dimType.Put("name", &nbt.StringTag{Value: "minecraft:the_void"})
	dimType.Put("id", &nbt.IntTag{Value: 0})

	elem := nbt.NewCompoundTag()
	elem.Put("piglin_safe", &nbt.ByteTag{Value: 0})
	elem.Put("natural", &nbt.ByteTag{Value: 0})
	elem.Put("ambient_light", &nbt.FloatTag{Value: 0.1})
	elem.Put("fixed_time", &nbt.IntArrayTag{Value: []int32{6000}})
	elem.Put("has_skylight", &nbt.ByteTag{Value: 0})
	elem.Put("has_ceiling", &nbt.ByteTag{Value: 0})
	elem.Put("ultrawarm", &nbt.ByteTag{Value: 0})
	elem.Put("has_raids", &nbt.ByteTag{Value: 0})
	elem.Put("respawn_anchor_works", &nbt.ByteTag{Value: 0})
	elem.Put("bed_works", &nbt.ByteTag{Value: 0})
	elem.Put("effects", &nbt.StringTag{Value: "minecraft:the_end"})
	elem.Put("min_y", &nbt.IntTag{Value: -64})
	elem.Put("height", &nbt.IntTag{Value: 384})
	elem.Put("logical_height", &nbt.IntTag{Value: 256})
	elem.Put("coordinate_scale", &nbt.DoubleTag{Value: 1})
	elem.Put("infiniburn", &nbt.StringTag{Value: "minecraft:infiniburn_end"})

	dimType.Put("element", elem)

	dimList := nbt.ListTag{Value: []nbt.Tag{dimType}}
	codec := nbt.NewCompoundTag()
	codec.Put("type", &nbt.StringTag{Value: "minecraft:dimension_type"})
	codec.Put("value", &dimList)

	root.Put("minecraft:dimension_type", codec)

	// Empty biome registry
	biomeCodec := nbt.NewCompoundTag()
	biomeCodec.Put("type", &nbt.StringTag{Value: "minecraft:worldgen/biome"})
	biomeCodec.Put("value", &nbt.ListTag{Value: []nbt.Tag{}})
	root.Put("minecraft:worldgen/biome", biomeCodec)

	var buf bytes.Buffer
	nbt.Write(&buf, nbt.NamedTag{Name: "", Tag: root}) // uncompressed
	return buf.Bytes()
}

// Main server logic
func main() {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", serverPort))
	if err != nil {
		panic(err)
	}
	fmt.Printf("Listening on 0.0.0.0:%d …\n", serverPort)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}

func handle(c net.Conn) {
	defer c.Close()
	state := 0 // 0: handshake, 1: status, 2: login

	for {
		length, err := readVarInt(c)
		if err != nil {
			return
		}
		pid, _ := readVarInt(c)
		data := make([]byte, length-1)
		if _, err = c.Read(data); err != nil {
			return
		}

		switch state {
		case 0: // handshake
			if len(data) > 0 {
				next := data[len(data)-1]
				if next == 1 {
					state = 1
				} else {
					state = 2
				}
			}

		case 1: // status
			if pid == 0 { // Status Request
				status := map[string]any{
					"version":     map[string]any{"name": "1.21.5", "protocol": protocolVer},
					"players":     map[string]any{"max": maxPlayers, "online": 0},
					"description": map[string]any{"text": motd},
				}
				js, _ := json.Marshal(status)
				var payload bytes.Buffer
				writeVarInt(&payload, int32(len(js)))
				payload.Write(js)
				sendPacket(c, 0, payload.Bytes()) // Response
			} else if pid == 1 { // Ping
				sendPacket(c, 1, data) // Pong
			}

		case 2: // login
			if pid == 0 { // Login Start
				// Read username (simplified parsing)
				nameLen := int(data[0])
				username := string(data[1 : 1+nameLen])

				fmt.Printf("User %s connecting...\n", username)

				// 1) Login Success
				var payload bytes.Buffer
				u := uuid.Nil
				payload.Write(u[:])
				writeString(&payload, username)
				sendPacket(c, 0x02, payload.Bytes())

				// 2) Move into Configuration state - send Join Game / registry
				reg := buildDimensionCodec()
				var join bytes.Buffer
				writeVarInt(&join, 0x00)                 // packet id in Configuration state
				writeVarInt(&join, 0)                    // dimension codec present as NBT
				join.Write(reg)                          // raw NBT
				writeString(&join, "minecraft:the_void") // dimension id
				writeVarInt(&join, 0)                    // hashed seed
				writeVarInt(&join, 0)                    // player gamemode (survival)
				writeVarInt(&join, -1)                   // previous gamemode (undefined)
				writeVarInt(&join, 0)                    // portal cooldown
				sendPacket(c, 0x00, join.Bytes())

				state = 3 // configuration/play; we never answer further
				fmt.Printf("User %s joined the void.\n", username)
				return // Close connection after join
			}
		}
	}
}
