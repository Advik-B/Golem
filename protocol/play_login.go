package protocol

import (
    "io"
)

// --- LoginPacket (Clientbound) ---
// This is the first packet sent in the Play state.

type LoginPacket struct {
    EntityID            int32
    IsHardcore          bool
    WorldNames          []string
    MaxPlayers          int32
    ViewDistance        int32
    SimulationDistance  int32
    ReducedDebugInfo    bool
    EnableRespawnScreen bool
    DoLimitedCrafting   bool
    WorldState          SpawnInfo
    EnforcesSecureChat  bool
}

func (p *LoginPacket) ID(version string) uint32                   { return 0x2A }
func (p *LoginPacket) State() State                               { return StatePlay }
func (p *LoginPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *LoginPacket) WriteTo(w io.Writer, version string) (err error) {
    if err = WriteInt32(w, p.EntityID); err != nil {
        return
    }
    if err = WriteBool(w, p.IsHardcore); err != nil {
        return
    }

    if err = WriteVarInt(w, int32(len(p.WorldNames))); err != nil {
        return
    }
    for _, name := range p.WorldNames {
        if err = WriteString(w, name); err != nil {
            return
        }
    }

    if err = WriteVarInt(w, p.MaxPlayers); err != nil {
        return
    }
    if err = WriteVarInt(w, p.ViewDistance); err != nil {
        return
    }
    if err = WriteVarInt(w, p.SimulationDistance); err != nil {
        return
    }
    if err = WriteBool(w, p.ReducedDebugInfo); err != nil {
        return
    }
    if err = WriteBool(w, p.EnableRespawnScreen); err != nil {
        return
    }
    if err = WriteBool(w, p.DoLimitedCrafting); err != nil {
        return
    }
    if err = p.WorldState.WriteTo(w); err != nil {
        return
    }
    err = WriteBool(w, p.EnforcesSecureChat)
    return
}

// --- RespawnPacket (Clientbound) ---

type RespawnPacket struct {
    WorldState   SpawnInfo
    CopyMetadata uint8 // Bitfield: 0x01 = Keep Attributes, 0x02 = Keep Entity Data
}

func (p *RespawnPacket) ID(version string) uint32                   { return 0x45 }
func (p *RespawnPacket) State() State                               { return StatePlay }
func (p *RespawnPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *RespawnPacket) WriteTo(w io.Writer, version string) (err error) {
    if err = p.WorldState.WriteTo(w); err != nil {
        return
    }
    err = WriteUint8(w, p.CopyMetadata)
    return
}

func init() {
    RegisterPacket(StatePlay, Clientbound, 0x2A, func() Packet { return &LoginPacket{} })
    RegisterPacket(StatePlay, Clientbound, 0x45, func() Packet { return &RespawnPacket{} })
}
