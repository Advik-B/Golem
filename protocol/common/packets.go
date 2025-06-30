package common

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/google/uuid"
)

func init() {
	// Common packets are registered in both configuration and play states
	states := []protocol.State{protocol.Configuration, protocol.Play}

	for _, state := range states {
		// Clientbound
		cb := protocol.Registry[state][protocol.Clientbound]
		cb.Register(ClientboundKeepAlivePacketID, func() protocol.Packet { return &ClientboundKeepAlivePacket{} })
		cb.Register(ClientboundPingPacketID, func() protocol.Packet { return &ClientboundPingPacket{} })
		cb.Register(ClientboundDisconnectPacketID, func() protocol.Packet { return &ClientboundDisconnectPacket{} })
		cb.Register(ClientboundResourcePackPushPacketID, func() protocol.Packet { return &ClientboundResourcePackPushPacket{} })
		// TODO: Register more common clientbound packets

		// Serverbound
		sb := protocol.Registry[state][protocol.Serverbound]
		sb.Register(ServerboundKeepAlivePacketID, func() protocol.Packet { return &ServerboundKeepAlivePacket{} })
		sb.Register(ServerboundPongPacketID, func() protocol.Packet { return &ServerboundPongPacket{} })
		sb.Register(ServerboundClientInformationID, func() protocol.Packet { return &ServerboundClientInformationPacket{} })
		// TODO: Register more common serverbound packets
	}
}

// Packet IDs for common packets.
const (
	// Clientbound (example IDs, actual values must be confirmed from wiki.vg for each state)
	ClientboundDisconnectPacketID       = 0x00 // Configuration state
	ClientboundKeepAlivePacketID        = 0x03 // Configuration state
	ClientboundPingPacketID             = 0x04 // Configuration state
	ClientboundResourcePackPushPacketID = 0x06 // Configuration state

	// Serverbound
	ServerboundClientInformationID = 0x00 // Configuration state
	ServerboundKeepAlivePacketID   = 0x02 // Configuration state
	ServerboundPongPacketID        = 0x03 // Configuration state
)

// --- Clientbound ---

type ClientboundKeepAlivePacket struct {
	KeepAliveID int64
}

func (pk *ClientboundKeepAlivePacket) ID() int32 { return ClientboundKeepAlivePacketID }
func (pk *ClientboundKeepAlivePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.KeepAliveID, err = r.ReadLong()
	return err
}
func (pk *ClientboundKeepAlivePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.KeepAliveID)
}

type ClientboundPingPacket struct {
	PingID int32
}

func (pk *ClientboundPingPacket) ID() int32 { return ClientboundPingPacketID }
func (pk *ClientboundPingPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.PingID, err = r.ReadInt()
	return err
}
func (pk *ClientboundPingPacket) WriteTo(w *codec.PacketBuffer) error { return w.WriteInt(pk.PingID) }

type ClientboundDisconnectPacket struct {
	Reason protocol.Component
}

func (pk *ClientboundDisconnectPacket) ID() int32 { return ClientboundDisconnectPacketID }
func (pk *ClientboundDisconnectPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.Reason.ReadFrom(r)
}
func (pk *ClientboundDisconnectPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.Reason.WriteTo(w)
}

type ClientboundResourcePackPushPacket struct {
	UUID     uuid.UUID // Corrected field name from ID to UUID
	URL      string
	Hash     string
	Required bool
	Prompt   *protocol.Component
}

func (pk *ClientboundResourcePackPushPacket) ID() int32 { return ClientboundResourcePackPushPacketID }
func (pk *ClientboundResourcePackPushPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.UUID, err = r.ReadUUID(); err != nil { // Correctly read into the new field name
		return err
	}
	if pk.URL, err = r.ReadString(32767); err != nil {
		return err
	}
	if pk.Hash, err = r.ReadString(40); err != nil {
		return err
	}
	if pk.Required, err = r.ReadBool(); err != nil {
		return err
	}
	hasPrompt, err := r.ReadBool()
	if err != nil {
		return err
	}
	if hasPrompt {
		var prompt protocol.Component
		if err := prompt.ReadFrom(r); err != nil {
			return err
		}
		pk.Prompt = &prompt
	}
	return nil
}
func (pk *ClientboundResourcePackPushPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteUUID(pk.UUID)
	w.WriteString(pk.URL)
	w.WriteString(pk.Hash)
	w.WriteBool(pk.Required)
	if pk.Prompt != nil {
		w.WriteBool(true)
		pk.Prompt.WriteTo(w)
	} else {
		w.WriteBool(false)
	}
	return nil
}

// --- Serverbound ---

type ServerboundKeepAlivePacket struct {
	KeepAliveID int64
}

func (pk *ServerboundKeepAlivePacket) ID() int32 { return ServerboundKeepAlivePacketID }
func (pk *ServerboundKeepAlivePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.KeepAliveID, err = r.ReadLong()
	return err
}
func (pk *ServerboundKeepAlivePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.KeepAliveID)
}

type ServerboundPongPacket struct {
	PongID int32
}

func (pk *ServerboundPongPacket) ID() int32 { return ServerboundPongPacketID }
func (pk *ServerboundPongPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.PongID, err = r.ReadInt()
	return err
}
func (pk *ServerboundPongPacket) WriteTo(w *codec.PacketBuffer) error { return w.WriteInt(pk.PongID) }

// ClientInformation contains settings from the client.
type ClientInformation struct {
	Locale              string
	ViewDistance        byte
	ChatMode            int32 // VarInt
	ChatColors          bool
	DisplayedSkinParts  byte
	MainHand            int32 // VarInt
	TextFiltering       bool
	AllowServerListings bool
}

func (ci *ClientInformation) ReadFrom(r *codec.PacketBuffer) (err error) {
	if ci.Locale, err = r.ReadString(16); err != nil {
		return
	}
	if ci.ViewDistance, err = r.ReadByte(); err != nil {
		return
	}
	if ci.ChatMode, err = r.ReadVarInt(); err != nil {
		return
	}
	if ci.ChatColors, err = r.ReadBool(); err != nil {
		return
	}
	if ci.DisplayedSkinParts, err = r.ReadByte(); err != nil {
		return
	}
	if ci.MainHand, err = r.ReadVarInt(); err != nil {
		return
	}
	if ci.TextFiltering, err = r.ReadBool(); err != nil {
		return
	}
	ci.AllowServerListings, err = r.ReadBool()
	return
}

func (ci *ClientInformation) WriteTo(w *codec.PacketBuffer) error {
	w.WriteString(ci.Locale)
	w.WriteByte(ci.ViewDistance)
	w.WriteVarInt(ci.ChatMode)
	w.WriteBool(ci.ChatColors)
	w.WriteByte(ci.DisplayedSkinParts)
	w.WriteVarInt(ci.MainHand)
	w.WriteBool(ci.TextFiltering)
	w.WriteBool(ci.AllowServerListings)
	return nil
}

type ServerboundClientInformationPacket struct {
	Info ClientInformation
}

func (pk *ServerboundClientInformationPacket) ID() int32 { return ServerboundClientInformationID }
func (pk *ServerboundClientInformationPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.Info.ReadFrom(r)
}
func (pk *ServerboundClientInformationPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.Info.WriteTo(w)
}
