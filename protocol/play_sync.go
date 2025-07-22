package protocol

import (
	"io"
)

// --- KeepAlivePacket (Clientbound) ---

type KeepAlivePacket struct {
	KeepAliveID int64
}

func (p *KeepAlivePacket) ID(version string) uint32                   { return 0x25 }
func (p *KeepAlivePacket) State() State                               { return StatePlay }
func (p *KeepAlivePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *KeepAlivePacket) WriteTo(w io.Writer, version string) (err error) {
	return WriteInt64(w, p.KeepAliveID)
}

// --- ServerKeepAlivePacket (Serverbound) ---

type ServerKeepAlivePacket struct {
	KeepAliveID int64
}

func (p *ServerKeepAlivePacket) ID(version string) uint32 { return 0x15 }
func (p *ServerKeepAlivePacket) State() State             { return StatePlay }
func (p *ServerKeepAlivePacket) ReadFrom(r io.Reader, version string) (err error) {
	p.KeepAliveID, err = ReadInt64(r)
	return
}
func (p *ServerKeepAlivePacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- PositionPacket (Clientbound) ---

type PositionPacket struct {
	TeleportID int32
	X, Y, Z    float64
	DX, DY, DZ float64
	Yaw, Pitch float32
	Flags      uint32
}

func (p *PositionPacket) ID(version string) uint32                   { return 0x3E }
func (p *PositionPacket) State() State                               { return StatePlay }
func (p *PositionPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *PositionPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.TeleportID); err != nil {
		return
	}
	if err = WriteFloat64(w, p.X); err != nil {
		return
	}
	if err = WriteFloat64(w, p.Y); err != nil {
		return
	}
	if err = WriteFloat64(w, p.Z); err != nil {
		return
	}
	if err = WriteFloat64(w, p.DX); err != nil {
		return
	}
	if err = WriteFloat64(w, p.DY); err != nil {
		return
	}
	if err = WriteFloat64(w, p.DZ); err != nil {
		return
	}
	if err = WriteFloat32(w, p.Yaw); err != nil {
		return
	}
	if err = WriteFloat32(w, p.Pitch); err != nil {
		return
	}
	return WriteUint32(w, p.Flags)
}

// --- PlayerPositionPacket (Serverbound) ---

type PlayerPositionPacket struct {
	X, Y, Z float64
	Flags   uint8
}

func (p *PlayerPositionPacket) ID(version string) uint32 { return 0x18 }
func (p *PlayerPositionPacket) State() State             { return StatePlay }
func (p *PlayerPositionPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.X, err = ReadFloat64(r); err != nil {
		return
	}
	if p.Y, err = ReadFloat64(r); err != nil {
		return
	}
	if p.Z, err = ReadFloat64(r); err != nil {
		return
	}
	p.Flags, err = ReadUint8(r)
	return
}
func (p *PlayerPositionPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- PlayerPositionRotationPacket (Serverbound) ---

type PlayerPositionRotationPacket struct {
	X, Y, Z    float64
	Yaw, Pitch float32
	Flags      uint8
}

func (p *PlayerPositionRotationPacket) ID(version string) uint32 { return 0x19 }
func (p *PlayerPositionRotationPacket) State() State             { return StatePlay }
func (p *PlayerPositionRotationPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.X, err = ReadFloat64(r); err != nil {
		return
	}
	if p.Y, err = ReadFloat64(r); err != nil {
		return
	}
	if p.Z, err = ReadFloat64(r); err != nil {
		return
	}
	if p.Yaw, err = ReadFloat32(r); err != nil {
		return
	}
	if p.Pitch, err = ReadFloat32(r); err != nil {
		return
	}
	p.Flags, err = ReadUint8(r)
	return
}
func (p *PlayerPositionRotationPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- PlayerRotationPacket (Serverbound) ---

type PlayerRotationPacket struct {
	Yaw, Pitch float32
	Flags      uint8
}

func (p *PlayerRotationPacket) ID(version string) uint32 { return 0x1A }
func (p *PlayerRotationPacket) State() State             { return StatePlay }
func (p *PlayerRotationPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Yaw, err = ReadFloat32(r); err != nil {
		return
	}
	if p.Pitch, err = ReadFloat32(r); err != nil {
		return
	}
	p.Flags, err = ReadUint8(r)
	return
}
func (p *PlayerRotationPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- PlayerOnGroundPacket (Serverbound) ---

type PlayerOnGroundPacket struct {
	Flags uint8
}

func (p *PlayerOnGroundPacket) ID(version string) uint32 { return 0x1B }
func (p *PlayerOnGroundPacket) State() State             { return StatePlay }
func (p *PlayerOnGroundPacket) ReadFrom(r io.Reader, version string) (err error) {
	p.Flags, err = ReadUint8(r)
	return
}
func (p *PlayerOnGroundPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- TeleportConfirmPacket (Serverbound) ---

type TeleportConfirmPacket struct {
	TeleportID int32
}

func (p *TeleportConfirmPacket) ID(version string) uint32 { return 0x00 }
func (p *TeleportConfirmPacket) State() State             { return StatePlay }
func (p *TeleportConfirmPacket) ReadFrom(r io.Reader, version string) (err error) {
	p.TeleportID, err = ReadVarInt(r)
	return
}
func (p *TeleportConfirmPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- UpdateTimePacket (Clientbound) ---

type UpdateTimePacket struct {
	WorldAge    int64
	TimeOfDay   int64
	TickDayTime bool
}

func (p *UpdateTimePacket) ID(version string) uint32                   { return 0x61 }
func (p *UpdateTimePacket) State() State                               { return StatePlay }
func (p *UpdateTimePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *UpdateTimePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteInt64(w, p.WorldAge); err != nil {
		return
	}
	if err = WriteInt64(w, p.TimeOfDay); err != nil {
		return
	}
	return WriteBool(w, p.TickDayTime)
}

func init() {
	// Clientbound
	RegisterPacket(StatePlay, Clientbound, 0x25, func() Packet { return &KeepAlivePacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x3E, func() Packet { return &PositionPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x61, func() Packet { return &UpdateTimePacket{} })

	// Serverbound
	RegisterPacket(StatePlay, Serverbound, 0x15, func() Packet { return &ServerKeepAlivePacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x18, func() Packet { return &PlayerPositionPacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x19, func() Packet { return &PlayerPositionRotationPacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x1A, func() Packet { return &PlayerRotationPacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x1B, func() Packet { return &PlayerOnGroundPacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x00, func() Packet { return &TeleportConfirmPacket{} })
}
