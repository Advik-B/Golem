package configuration

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec" // Corrected import
	"github.com/Advik-B/Golem/protocol/common"
)

func init() {
	sb := protocol.Registry[protocol.Configuration][protocol.Serverbound]
	sb.Register(common.ServerboundClientInformationID, func() protocol.Packet { return &common.ServerboundClientInformationPacket{} })
	// Cookie Response is registered in the cookie package's init
	sb.Register(common.ServerboundCustomPayloadID, func() protocol.Packet { return &common.ServerboundCustomPayloadPacket{} })
	sb.Register(common.ServerboundFinishConfigurationID, func() protocol.Packet { return &ServerboundFinishConfigurationPacket{} })
	sb.Register(common.ServerboundKeepAliveID, func() protocol.Packet { return &common.ServerboundKeepAlivePacket{} })
	sb.Register(common.ServerboundPongID, func() protocol.Packet { return &common.ServerboundPongPacket{} })
	sb.Register(common.ServerboundResourcePackID, func() protocol.Packet { return &common.ServerboundResourcePackPacket{} })

	cb := protocol.Registry[protocol.Configuration][protocol.Clientbound]
	cb.Register(ClientboundFinishConfigurationPacketID, func() protocol.Packet { return &ClientboundFinishConfigurationPacket{} })
	// TODO: Register other clientbound config packets
}

// Packet IDs specific to Configuration or that differ from Play state.
// Using correct IDs per wiki.vg for 1.21 Configuration State
const (
	ClientboundFinishConfigurationPacketID = 0x02
)

// ServerboundFinishConfigurationPacket is unique to this state.
type ServerboundFinishConfigurationPacket struct{}

func (pk *ServerboundFinishConfigurationPacket) ID() int32 {
	return common.ServerboundFinishConfigurationID
}
func (pk *ServerboundFinishConfigurationPacket) ReadFrom(r *codec.PacketBuffer) error {
	return nil
}
func (pk *ServerboundFinishConfigurationPacket) WriteTo(w *codec.PacketBuffer) error {
	return nil
}

// ClientboundFinishConfigurationPacket is unique to this state.
type ClientboundFinishConfigurationPacket struct{}

func (pk *ClientboundFinishConfigurationPacket) ID() int32 {
	return ClientboundFinishConfigurationPacketID
}
func (pk *ClientboundFinishConfigurationPacket) ReadFrom(r *codec.PacketBuffer) error {
	return nil
}
func (pk *ClientboundFinishConfigurationPacket) WriteTo(w *codec.PacketBuffer) error {
	return nil
}
