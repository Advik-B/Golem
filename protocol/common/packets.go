package common

import (
	"github.com/Advik-B/Golem/protocol"
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/google/uuid"
)

// This file defines packets that are shared across multiple protocol states,
// primarily between Configuration and Play.

// Packet IDs for common packets.
const (
	// Clientbound
	ClientboundKeepAlivePacketID        = 0x24 // Example ID, will use correct ones
	ClientboundPingPacketID             = 0x30
	ClientboundDisconnectPacketID       = 0x1A
	ClientboundResourcePackPushPacketID = 0x38

	// Serverbound
	ServerboundKeepAlivePacketID   = 0x14
	ServerboundPongPacketID        = 0x1D
	ServerboundClientInformationID = 0x08
)

// --- Clientbound ---

type ClientboundKeepAlivePacket struct {
	ID int64
}

func (pk *ClientboundKeepAlivePacket) ID() int32 { return ClientboundKeepAlivePacketID }
func (pk *ClientboundKeepAlivePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.ID, err = r.ReadLong()
	return err
}
func (pk *ClientboundKeepAlivePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.ID)
}

type ClientboundPingPacket struct {
	ID int32
}

func (pk *ClientboundPingPacket) ID() int32 { return ClientboundPingPacketID }
func (pk *ClientboundPingPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.ID, err = r.ReadInt()
	return err
}
func (pk *ClientboundPingPacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteInt(pk.ID)
}

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
	ID       uuid.UUID
	URL      string
	Hash     string
	Required bool
	Prompt   *protocol.Component
}

func (pk *ClientboundResourcePackPushPacket) ID() int32 { return ClientboundResourcePackPushPacketID }
func (pk *ClientboundResourcePackPushPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if pk.ID, err = r.ReadUUID(); err != nil {
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
	w.WriteUUID(pk.ID)
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
	ID int64
}

func (pk *ServerboundKeepAlivePacket) ID() int32 { return ServerboundKeepAlivePacketID }
func (pk *ServerboundKeepAlivePacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.ID, err = r.ReadLong()
	return err
}
func (pk *ServerboundKeepAlivePacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteLong(pk.ID)
}

type ServerboundPongPacket struct {
	ID int32
}

func (pk *ServerboundPongPacket) ID() int32 { return ServerboundPongPacketID }
func (pk *ServerboundPongPacket) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	pk.ID, err = r.ReadInt()
	return err
}
func (pk *ServerboundPongPacket) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteInt(pk.ID)
}

type ServerboundClientInformationPacket struct {
	// Omitting fields for brevity for now. This would include locale, view distance, etc.
}

func (pk *ServerboundClientInformationPacket) ID() int32                            { return ServerboundClientInformationID }
func (pk *ServerboundClientInformationPacket) ReadFrom(r *codec.PacketBuffer) error { return nil } // TODO
func (pk *ServerboundClientInformationPacket) WriteTo(w *codec.PacketBuffer) error  { return nil } // TODO
