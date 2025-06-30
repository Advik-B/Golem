package cookie

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
)

func init() {
	// Cookie packets can be sent during Login and Configuration
	states := []protocol.State{protocol.Login, protocol.Configuration}
	for _, state := range states {
		sb := protocol.Registry[state][protocol.Serverbound]
		sb.Register(ServerboundCookieResponsePacketID, func() protocol.Packet { return &ServerboundCookieResponsePacket{} })

		cb := protocol.Registry[state][protocol.Clientbound]
		cb.Register(ClientboundCookieRequestPacketID, func() protocol.Packet { return &ClientboundCookieRequestPacket{} })
	}
}

const (
	ServerboundCookieResponsePacketID = 0x01
	ClientboundCookieRequestPacketID  = 0x01
)

type ServerboundCookieResponsePacket struct {
	Key     protocol.ResourceLocation
	Payload []byte // Optional
}

func (pk *ServerboundCookieResponsePacket) ID() int32 { return ServerboundCookieResponsePacketID }
func (pk *ServerboundCookieResponsePacket) ReadFrom(r *codec.PacketBuffer) (err error) {
	if err = pk.Key.ReadFrom(r); err != nil {
		return
	}
	hasPayload, err := r.ReadBool()
	if err != nil || !hasPayload {
		return err
	}
	pk.Payload, err = r.ReadByteArray()
	return
}
func (pk *ServerboundCookieResponsePacket) WriteTo(w *codec.PacketBuffer) error {
	pk.Key.WriteTo(w)
	hasPayload := pk.Payload != nil
	w.WriteBool(hasPayload)
	if hasPayload {
		w.WriteByteArray(pk.Payload)
	}
	return nil
}

type ClientboundCookieRequestPacket struct {
	Key protocol.ResourceLocation
}

func (pk *ClientboundCookieRequestPacket) ID() int32 { return ClientboundCookieRequestPacketID }
func (pk *ClientboundCookieRequestPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.Key.ReadFrom(r)
}
func (pk *ClientboundCookieRequestPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.Key.WriteTo(w)
}
