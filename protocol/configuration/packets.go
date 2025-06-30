package configuration

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
)

func init() {
	sb := protocol.Registry[protocol.Configuration][protocol.Serverbound]
	sb.Register(ServerboundFinishConfigurationPacketID, func() protocol.Packet { return &ServerboundFinishConfigurationPacket{} })
	sb.Register(ServerboundSelectKnownPacksPacketID, func() protocol.Packet { return &ServerboundSelectKnownPacksPacket{} })

	cb := protocol.Registry[protocol.Configuration][protocol.Clientbound]
	cb.Register(ClientboundFinishConfigurationPacketID, func() protocol.Packet { return &ClientboundFinishConfigurationPacket{} })
	cb.Register(ClientboundSelectKnownPacksPacketID, func() protocol.Packet { return &ClientboundSelectKnownPacksPacket{} })
	// TODO: Register RegistryData, UpdateEnabledFeatures
}

const (
	// Serverbound
	ServerboundSelectKnownPacksPacketID    = 0x01
	ServerboundFinishConfigurationPacketID = 0x02

	// Clientbound
	ClientboundSelectKnownPacksPacketID    = 0x04
	ClientboundFinishConfigurationPacketID = 0x02
)

// KnownPack is a simplified version for this state.
type KnownPack struct {
	Namespace string
	ID        string
	Version   string
}

func (kp *KnownPack) ReadFrom(r *codec.PacketBuffer) (err error) {
	if kp.Namespace, err = r.ReadString(32767); err != nil {
		return
	}
	if kp.ID, err = r.ReadString(32767); err != nil {
		return
	}
	kp.Version, err = r.ReadString(32767)
	return
}

func (kp *KnownPack) WriteTo(w *codec.PacketBuffer) error {
	w.WriteString(kp.Namespace)
	w.WriteString(kp.ID)
	w.WriteString(kp.Version)
	return nil
}

// --- Clientbound ---

type ClientboundSelectKnownPacksPacket struct {
	KnownPacks []KnownPack
}

func (pk *ClientboundSelectKnownPacksPacket) ID() int32 { return ClientboundSelectKnownPacksPacketID }
func (pk *ClientboundSelectKnownPacksPacket) ReadFrom(r *codec.PacketBuffer) error {
	count, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	pk.KnownPacks = make([]KnownPack, count)
	for i := int32(0); i < count; i++ {
		if err := pk.KnownPacks[i].ReadFrom(r); err != nil {
			return err
		}
	}
	return nil
}
func (pk *ClientboundSelectKnownPacksPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteVarInt(int32(len(pk.KnownPacks)))
	for _, pack := range pk.KnownPacks {
		if err := pack.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}

type ClientboundFinishConfigurationPacket struct{}

func (pk *ClientboundFinishConfigurationPacket) ID() int32 {
	return ClientboundFinishConfigurationPacketID
}
func (pk *ClientboundFinishConfigurationPacket) ReadFrom(r *codec.PacketBuffer) error { return nil }
func (pk *ClientboundFinishConfigurationPacket) WriteTo(w *codec.PacketBuffer) error  { return nil }

// --- Serverbound ---

type ServerboundSelectKnownPacksPacket struct {
	KnownPacks []KnownPack
}

func (pk *ServerboundSelectKnownPacksPacket) ID() int32 { return ServerboundSelectKnownPacksPacketID }
func (pk *ServerboundSelectKnownPacksPacket) ReadFrom(r *codec.PacketBuffer) error {
	count, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	pk.KnownPacks = make([]KnownPack, count)
	for i := int32(0); i < count; i++ {
		if err := pk.KnownPacks[i].ReadFrom(r); err != nil {
			return err
		}
	}
	return nil
}
func (pk *ServerboundSelectKnownPacksPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteVarInt(int32(len(pk.KnownPacks)))
	for _, pack := range pk.KnownPacks {
		if err := pack.WriteTo(w); err != nil {
			return err
		}
	}
	return nil
}

type ServerboundFinishConfigurationPacket struct{}

func (pk *ServerboundFinishConfigurationPacket) ID() int32 {
	return ServerboundFinishConfigurationPacketID
}
func (pk *ServerboundFinishConfigurationPacket) ReadFrom(r *codec.PacketBuffer) error { return nil }
func (pk *ServerboundFinishConfigurationPacket) WriteTo(w *codec.PacketBuffer) error  { return nil }
