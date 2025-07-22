package protocol

import (
	"io"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

// --- PlayerChatPacket (Clientbound) ---

type PlayerChatPacket struct {
	SenderUUID       uuid.UUID
	Index            int32
	Signature        *[]byte // Optional
	PlainMessage     string
	Timestamp        int64
	Salt             int64
	PreviousMessages []struct { // Simplified for now
		ID        int32
		Signature *[]byte
	}
	UnsignedChatContent *nbt.Tag
	FilterType          int32
	FilterTypeMask      []int64
	Type                int32 // ChatTypesHolder
	NetworkName         nbt.Tag
	NetworkTargetName   *nbt.Tag
}

func (p *PlayerChatPacket) ID(version string) uint32                   { return 0x3D }
func (p *PlayerChatPacket) State() State                               { return StatePlay }
func (p *PlayerChatPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *PlayerChatPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteUUID(w, p.SenderUUID); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Index); err != nil {
		return
	}

	hasSig := p.Signature != nil
	if err = WriteBool(w, hasSig); err != nil {
		return
	}
	if hasSig {
		if _, err = w.Write(*p.Signature); err != nil {
			return
		}
	}

	if err = WriteString(w, p.PlainMessage); err != nil {
		return
	}
	if err = WriteInt64(w, p.Timestamp); err != nil {
		return
	}
	if err = WriteInt64(w, p.Salt); err != nil {
		return
	}

	if err = WriteVarInt(w, int32(len(p.PreviousMessages))); err != nil {
		return
	}
	for _, msg := range p.PreviousMessages {
		if err = WriteVarInt(w, msg.ID); err != nil {
			return
		}
		if msg.ID == 0 { // Has signature
			if _, err = w.Write(*msg.Signature); err != nil {
				return
			}
		}
	}

	hasUnsigned := p.UnsignedChatContent != nil
	if err = WriteBool(w, hasUnsigned); err != nil {
		return
	}
	if hasUnsigned {
		if err = WriteNBT(w, *p.UnsignedChatContent); err != nil {
			return
		}
	}

	if err = WriteVarInt(w, p.FilterType); err != nil {
		return
	}
	if p.FilterType == 2 {
		if err = WriteVarInt(w, int32(len(p.FilterTypeMask))); err != nil {
			return
		}
		for _, mask := range p.FilterTypeMask {
			if err = WriteInt64(w, mask); err != nil {
				return
			}
		}
	}

	if err = WriteVarInt(w, p.Type); err != nil {
		return
	}
	if err = WriteNBT(w, p.NetworkName); err != nil {
		return
	}

	hasTarget := p.NetworkTargetName != nil
	if err = WriteBool(w, hasTarget); err != nil {
		return
	}
	if hasTarget {
		err = WriteNBT(w, *p.NetworkTargetName)
	}

	return
}

// --- SystemChatPacket (Clientbound) ---

type SystemChatPacket struct {
	Content     nbt.Tag
	IsActionBar bool
}

func (p *SystemChatPacket) ID(version string) uint32                   { return 0x66 }
func (p *SystemChatPacket) State() State                               { return StatePlay }
func (p *SystemChatPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *SystemChatPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteNBT(w, p.Content); err != nil {
		return
	}
	return WriteBool(w, p.IsActionBar)
}

// --- TabCompletePacket (Clientbound) ---

type TabCompletePacket struct {
	TransactionID int32
	Start         int32
	Length        int32
	Matches       []struct {
		Match   string
		Tooltip *nbt.Tag
	}
}

func (p *TabCompletePacket) ID(version string) uint32                   { return 0x11 }
func (p *TabCompletePacket) State() State                               { return StatePlay }
func (p *TabCompletePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *TabCompletePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.TransactionID); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Start); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Length); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Matches))); err != nil {
		return
	}
	for _, match := range p.Matches {
		if err = WriteString(w, match.Match); err != nil {
			return
		}
		hasTooltip := match.Tooltip != nil
		if err = WriteBool(w, hasTooltip); err != nil {
			return
		}
		if hasTooltip {
			if err = WriteNBT(w, *match.Tooltip); err != nil {
				return
			}
		}
	}
	return
}

// --- ChatMessagePacket (Serverbound) ---

type ChatMessagePacket struct {
	Message      string
	Timestamp    int64
	Salt         int64
	Signature    *[]byte
	Offset       int32
	Acknowledged [3]byte
}

func (p *ChatMessagePacket) ID(version string) uint32 { return 0x07 }
func (p *ChatMessagePacket) State() State             { return StatePlay }
func (p *ChatMessagePacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.Message, err = ReadString(r); err != nil {
		return
	}
	if p.Timestamp, err = ReadInt64(r); err != nil {
		return
	}
	if p.Salt, err = ReadInt64(r); err != nil {
		return
	}

	hasSig, err := ReadBool(r)
	if err != nil {
		return
	}
	if hasSig {
		sig := make([]byte, 256)
		if _, err = io.ReadFull(r, sig); err != nil {
			return
		}
		p.Signature = &sig
	}

	if p.Offset, err = ReadVarInt(r); err != nil {
		return
	}
	_, err = io.ReadFull(r, p.Acknowledged[:])
	return
}
func (p *ChatMessagePacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

// --- TabCompleteRequestPacket (Serverbound) ---

type TabCompleteRequestPacket struct {
	TransactionID int32
	Text          string
}

func (p *TabCompleteRequestPacket) ID(version string) uint32 { return 0x0A }
func (p *TabCompleteRequestPacket) State() State             { return StatePlay }
func (p *TabCompleteRequestPacket) ReadFrom(r io.Reader, version string) (err error) {
	if p.TransactionID, err = ReadVarInt(r); err != nil {
		return
	}
	p.Text, err = ReadString(r)
	return
}
func (p *TabCompleteRequestPacket) WriteTo(w io.Writer, version string) error { return nil } // Client only

func init() {
	// Clientbound
	RegisterPacket(StatePlay, Clientbound, 0x3D, func() Packet { return &PlayerChatPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x66, func() Packet { return &SystemChatPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x11, func() Packet { return &TabCompletePacket{} })

	// Serverbound
	RegisterPacket(StatePlay, Serverbound, 0x07, func() Packet { return &ChatMessagePacket{} })
	RegisterPacket(StatePlay, Serverbound, 0x0A, func() Packet { return &TabCompleteRequestPacket{} })
}
