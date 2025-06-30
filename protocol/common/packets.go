package common

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/google/uuid"
)

// This file defines packets that are shared across multiple protocol states,
// primarily between Configuration and Play. It does NOT register them.

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

// --- Common Packets ---

// Constants are for reference. Registration happens in state-specific packages.
const (
	ServerboundClientInformationID   = 0x00
	ServerboundCustomPayloadID       = 0x01 // Correct ID for Configuration state
	ServerboundFinishConfigurationID = 0x02 // Correct ID for Configuration state
	ServerboundKeepAliveID           = 0x03 // Correct ID for Configuration state
	ServerboundPongID                = 0x04 // Correct ID for Configuration state
	ServerboundResourcePackID        = 0x05 // Correct ID for Configuration state
)

type ServerboundClientInformationPacket struct{ Info ClientInformation }

func (pk *ServerboundClientInformationPacket) ID() int32 { return ServerboundClientInformationID }
func (pk *ServerboundClientInformationPacket) ReadFrom(r *codec.PacketBuffer) error {
	return pk.Info.ReadFrom(r)
}
func (pk *ServerboundClientInformationPacket) WriteTo(w *codec.PacketBuffer) error {
	return pk.Info.WriteTo(w)
}

type ServerboundKeepAlivePacket struct{ KeepAliveID int64 }

func (pk *ServerboundKeepAlivePacket) ID() int32 { return ServerboundKeepAliveID }
func (pk *ServerboundKeepAlivePacket) ReadFrom(r *codec.PacketBuffer) error {
	pk.KeepAliveID, _ = r.ReadLong()
	return nil
}
func (pk *ServerboundKeepAlivePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.KeepAliveID)
}

type ServerboundPongPacket struct{ PongID int32 }

func (pk *ServerboundPongPacket) ID() int32 { return ServerboundPongID }
func (pk *ServerboundPongPacket) ReadFrom(r *codec.PacketBuffer) error {
	pk.PongID, _ = r.ReadInt()
	return nil
}
func (pk *ServerboundPongPacket) WriteTo(w *codec.PacketBuffer) error { return w.WriteInt(pk.PongID) }

// NEW: Added ServerboundCustomPayloadPacket
type ServerboundCustomPayloadPacket struct {
	Identifier protocol.ResourceLocation
	Data       []byte
}

func (pk *ServerboundCustomPayloadPacket) ID() int32 { return ServerboundCustomPayloadID }
func (pk *ServerboundCustomPayloadPacket) ReadFrom(r *codec.PacketBuffer) (err error) {
	if err = pk.Identifier.ReadFrom(r); err != nil {
		return
	}
	// The rest of the buffer is data.
	pk.Data = make([]byte, r.Len())
	_, err = r.Read(pk.Data)
	return
}
func (pk *ServerboundCustomPayloadPacket) WriteTo(w *codec.PacketBuffer) error {
	pk.Identifier.WriteTo(w)
	w.Write(pk.Data)
	return nil
}

// ResourcePackStatus represents the result of a client resource pack load.
type ResourcePackStatus int32

const (
	SuccessfullyLoaded ResourcePackStatus = iota
	Declined
	FailedDownload
	Accepted
)

// NEW: Added ServerboundResourcePackPacket
type ServerboundResourcePackPacket struct {
	UUID   uuid.UUID
	Status ResourcePackStatus
}

func (pk *ServerboundResourcePackPacket) ID() int32 { return ServerboundResourcePackID }
func (pk *ServerboundResourcePackPacket) ReadFrom(r *codec.PacketBuffer) (err error) {
	if pk.UUID, err = r.ReadUUID(); err != nil {
		return
	}
	status, err := r.ReadVarInt()
	if err != nil {
		return err
	}
	pk.Status = ResourcePackStatus(status)
	return
}
func (pk *ServerboundResourcePackPacket) WriteTo(w *codec.PacketBuffer) error {
	w.WriteUUID(pk.UUID)
	w.WriteVarInt(int32(pk.Status))
	return nil
}
