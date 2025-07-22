package protocol

import (
    "io"
)

// --- ServerInfoPacket (Clientbound) ---
// Also known as "Response" packet.

type ServerInfoPacket struct {
    Response string // JSON encoded server status
}

func (p *ServerInfoPacket) ID(version string) uint32 { return 0x00 }
func (p *ServerInfoPacket) State() State             { return StateStatus }
func (p *ServerInfoPacket) ReadFrom(r io.Reader, version string) (err error) {
    p.Response, err = ReadString(r)
    return
}
func (p *ServerInfoPacket) WriteTo(w io.Writer, version string) (err error) {
    return WriteString(w, p.Response)
}

// --- PingPacket (Clientbound & Serverbound) ---
// The name is slightly confusing. The client sends a "Ping" (or "PingRequest")
// and the server responds with a "Pong" (or "PongResponse"). In minecraft-protocol,
// both use the same name and structure. We'll follow that convention.
// Serverbound is ID 0x01, Clientbound is ID 0x01.

type PingPacket struct {
    Payload int64
}

func (p *PingPacket) ID(version string) uint32 { return 0x01 }
func (p *PingPacket) State() State             { return StateStatus }
func (p *PingPacket) ReadFrom(r io.Reader, version string) (err error) {
    p.Payload, err = ReadInt64(r)
    return
}
func (p *PingPacket) WriteTo(w io.Writer, version string) (err error) {
    return WriteInt64(w, p.Payload)
}

// --- PingStartPacket (Serverbound) ---
// Also known as "Request" packet.

type PingStartPacket struct {
    // This packet has no fields.
}

func (p *PingStartPacket) ID(version string) uint32 { return 0x00 }
func (p *PingStartPacket) State() State             { return StateStatus }
func (p *PingStartPacket) ReadFrom(r io.Reader, version string) (err error) {
    return nil // No fields to read
}
func (p *PingStartPacket) WriteTo(w io.Writer, version string) (err error) {
    return nil // No fields to write
}

// init registers the packets in this file.
func init() {
    // Clientbound (Server -> Client)
    RegisterPacket(StateStatus, Clientbound, 0x00, func() Packet { return &ServerInfoPacket{} })
    RegisterPacket(StateStatus, Clientbound, 0x01, func() Packet { return &PingPacket{} })

    // Serverbound (Client -> Server)
    RegisterPacket(StateStatus, Serverbound, 0x00, func() Packet { return &PingStartPacket{} })
    RegisterPacket(StateStatus, Serverbound, 0x01, func() Packet { return &PingPacket{} })
}
