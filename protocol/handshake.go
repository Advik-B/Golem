package protocol

import (
    "io"
)

// --- SetProtocolPacket (Serverbound) ---

type SetProtocolPacket struct {
    ProtocolVersion int32
    ServerHost      string
    ServerPort      uint16
    NextState       int32
}

func (p *SetProtocolPacket) ID(version string) uint32 { return 0x00 }
func (p *SetProtocolPacket) State() State             { return StateHandshaking }
func (p *SetProtocolPacket) ReadFrom(r io.Reader, version string) (err error) {
    if p.ProtocolVersion, err = ReadVarInt(r); err != nil {
        return
    }
    if p.ServerHost, err = ReadString(r); err != nil {
        return
    }
    if p.ServerPort, err = ReadUint16(r); err != nil {
        return
    }
    if p.NextState, err = ReadVarInt(r); err != nil {
        return
    }
    return
}
func (p *SetProtocolPacket) WriteTo(w io.Writer, version string) (err error) {
    if err = WriteVarInt(w, p.ProtocolVersion); err != nil {
        return
    }
    if err = WriteString(w, p.ServerHost); err != nil {
        return
    }
    if err = WriteUint16(w, p.ServerPort); err != nil {
        return
    }
    if err = WriteVarInt(w, p.NextState); err != nil {
        return
    }
    return
}

// --- LegacyServerListPingPacket (Serverbound) ---
// This packet is special and has a minimal payload.
type LegacyServerListPingPacket struct {
    Payload byte
}

func (p *LegacyServerListPingPacket) ID(version string) uint32 { return 0xFE }
func (p *LegacyServerListPingPacket) State() State             { return StateHandshaking }
func (p *LegacyServerListPingPacket) ReadFrom(r io.Reader, version string) (err error) {
    // The payload is typically checked, but for a simple server it can be ignored.
    // We read one byte to satisfy the protocol.
    p.Payload, err = ReadUint8(r)
    return
}
func (p *LegacyServerListPingPacket) WriteTo(w io.Writer, version string) (err error) {
    return WriteUint8(w, p.Payload)
}

// init registers the packets in this file.
func init() {
    RegisterPacket(StateHandshaking, Serverbound, 0x00, func() Packet { return &SetProtocolPacket{} })
    RegisterPacket(StateHandshaking, Serverbound, 0xFE, func() Packet { return &LegacyServerListPingPacket{} })
}
