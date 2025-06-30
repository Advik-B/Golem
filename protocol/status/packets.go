package status

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
)

func init() {
	// Serverbound
	sb := protocol.Registry[protocol.Status][protocol.Serverbound]
	sb.Register(ServerboundStatusRequestPacketID, func() protocol.Packet { return &ServerboundStatusRequestPacket{} })
	sb.Register(ServerboundPingRequestPacketID, func() protocol.Packet { return &ServerboundPingRequestPacket{} })

	// Clientbound
	cb := protocol.Registry[protocol.Status][protocol.Clientbound]
	cb.Register(ClientboundStatusResponsePacketID, func() protocol.Packet { return &ClientboundStatusResponsePacket{} })
	cb.Register(ClientboundPongResponsePacketID, func() protocol.Packet { return &ClientboundPongResponsePacket{} })
}

const (
	// Serverbound
	ServerboundStatusRequestPacketID = 0x00
	ServerboundPingRequestPacketID   = 0x01

	// Clientbound
	ClientboundStatusResponsePacketID = 0x00
	ClientboundPongResponsePacketID   = 0x01
)

// --- Serverbound ---

type ServerboundStatusRequestPacket struct{}

func (pk *ServerboundStatusRequestPacket) ID() int32                            { return ServerboundStatusRequestPacketID }
func (pk *ServerboundStatusRequestPacket) ReadFrom(r *codec.PacketBuffer) error { return nil }
func (pk *ServerboundStatusRequestPacket) WriteTo(w *codec.PacketBuffer) error  { return nil }

type ServerboundPingRequestPacket struct {
	Time int64
}

func (pk *ServerboundPingRequestPacket) ID() int32 { return ServerboundPingRequestPacketID }
func (pk *ServerboundPingRequestPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.Time, err = r.ReadLong()
	return err
}
func (pk *ServerboundPingRequestPacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.Time)
}

// --- Clientbound ---

type ClientboundStatusResponsePacket struct {
	StatusJSON string // ServerStatus will be handled by the server logic
}

func (pk *ClientboundStatusResponsePacket) ID() int32 { return ClientboundStatusResponsePacketID }
func (pk *ClientboundStatusResponsePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.StatusJSON, err = r.ReadString(32767)
	return err
}
func (pk *ClientboundStatusResponsePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteString(pk.StatusJSON)
}

type ClientboundPongResponsePacket struct {
	Time int64
}

func (pk *ClientboundPongResponsePacket) ID() int32 { return ClientboundPongResponsePacketID }
func (pk *ClientboundPongResponsePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.Time, err = r.ReadLong()
	return err
}
func (pk *ClientboundPongResponsePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.Time)
}
