package protocol

import (
	"io"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

// --- SpawnEntityPacket (Clientbound) ---

type SpawnEntityPacket struct {
	EntityID         int32
	ObjectUUID       uuid.UUID
	Type             int32
	X, Y, Z          float64
	Pitch, Yaw       int8
	HeadPitch        int8
	ObjectData       int32
	VelX, VelY, VelZ int16
}

func (p *SpawnEntityPacket) ID(version string) uint32                   { return 0x01 }
func (p *SpawnEntityPacket) State() State                               { return StatePlay }
func (p *SpawnEntityPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *SpawnEntityPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteUUID(w, p.ObjectUUID); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Type); err != nil {
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
	if err = WriteInt8(w, p.Pitch); err != nil {
		return
	}
	if err = WriteInt8(w, p.Yaw); err != nil {
		return
	}
	if err = WriteInt8(w, p.HeadPitch); err != nil {
		return
	}
	if err = WriteVarInt(w, p.ObjectData); err != nil {
		return
	}
	if err = WriteInt16(w, p.VelX); err != nil {
		return
	}
	if err = WriteInt16(w, p.VelY); err != nil {
		return
	}
	return WriteInt16(w, p.VelZ)
}

// --- SpawnExperienceOrbPacket (Clientbound) ---

type SpawnExperienceOrbPacket struct {
	EntityID int32
	X, Y, Z  float64
	Count    int16
}

func (p *SpawnExperienceOrbPacket) ID(version string) uint32                   { return 0x02 }
func (p *SpawnExperienceOrbPacket) State() State                               { return StatePlay }
func (p *SpawnExperienceOrbPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *SpawnExperienceOrbPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
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
	return WriteInt16(w, p.Count)
}

// --- PlayerInfoUpdatePacket (Clientbound) ---
type PlayerInfoUpdatePacket struct {
	ActionFlags uint8
	Players     []PlayerInfoUpdateData
}

type PlayerInfoUpdateData struct {
	UUID         uuid.UUID
	Player       *GameProfile
	ChatSession  *ChatSession
	GameMode     int32
	Listed       int32
	Latency      int32
	DisplayName  *nbt.Tag
	ListPriority int32
	ShowHat      bool
}

type GameProfile struct {
	Name       string
	Properties []struct {
		Name      string
		Value     string
		Signature *string
	}
}

func (p *PlayerInfoUpdatePacket) ID(version string) uint32                   { return 0x40 }
func (p *PlayerInfoUpdatePacket) State() State                               { return StatePlay }
func (p *PlayerInfoUpdatePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *PlayerInfoUpdatePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteUint8(w, p.ActionFlags); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Players))); err != nil {
		return
	}

	for _, data := range p.Players {
		if err = WriteUUID(w, data.UUID); err != nil {
			return
		}

		if (p.ActionFlags & 0x01) != 0 { // add_player
			if err = WriteString(w, data.Player.Name); err != nil {
				return
			}
			if err = WriteVarInt(w, int32(len(data.Player.Properties))); err != nil {
				return
			}
			for _, prop := range data.Player.Properties {
				if err = WriteString(w, prop.Name); err != nil {
					return
				}
				if err = WriteString(w, prop.Value); err != nil {
					return
				}
				hasSig := prop.Signature != nil
				if err = WriteBool(w, hasSig); err != nil {
					return
				}
				if hasSig {
					if err = WriteString(w, *prop.Signature); err != nil {
						return
					}
				}
			}
		}

		if (p.ActionFlags & 0x02) != 0 { // initialize_chat
			hasChat := data.ChatSession != nil
			if err = WriteBool(w, hasChat); err != nil {
				return
			}
			if hasChat {
				// write chat session data
			}
		}

		if (p.ActionFlags & 0x04) != 0 { // update_game_mode
			if err = WriteVarInt(w, data.GameMode); err != nil {
				return
			}
		}

		if (p.ActionFlags & 0x08) != 0 { // update_listed
			if err = WriteVarInt(w, data.Listed); err != nil {
				return
			}
		}

		if (p.ActionFlags & 0x10) != 0 { // update_latency
			if err = WriteVarInt(w, data.Latency); err != nil {
				return
			}
		}

		if (p.ActionFlags & 0x20) != 0 { // update_display_name
			hasDisplay := data.DisplayName != nil
			if err = WriteBool(w, hasDisplay); err != nil {
				return
			}
			if hasDisplay {
				if err = WriteNBT(w, *data.DisplayName); err != nil {
					return
				}
			}
		}

		if (p.ActionFlags & 0x40) != 0 { // update_hat
			if err = WriteBool(w, data.ShowHat); err != nil {
				return
			}
		}

		if (p.ActionFlags & 0x80) != 0 { // update_list_order
			if err = WriteVarInt(w, data.ListPriority); err != nil {
				return
			}
		}
	}
	return
}

// --- PlayerInfoRemovePacket (Clientbound) ---
type PlayerInfoRemovePacket struct {
	Players []uuid.UUID
}

func (p *PlayerInfoRemovePacket) ID(version string) uint32                   { return 0x3F }
func (p *PlayerInfoRemovePacket) State() State                               { return StatePlay }
func (p *PlayerInfoRemovePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *PlayerInfoRemovePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, int32(len(p.Players))); err != nil {
		return
	}
	for _, u := range p.Players {
		if err = WriteUUID(w, u); err != nil {
			return
		}
	}
	return
}

// --- EntityTeleportPacket (Clientbound) ---

type EntityTeleportPacket struct {
	EntityID   int32
	X, Y, Z    float64
	Yaw, Pitch int8
	OnGround   bool
}

func (p *EntityTeleportPacket) ID(version string) uint32                   { return 0x6C }
func (p *EntityTeleportPacket) State() State                               { return StatePlay }
func (p *EntityTeleportPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityTeleportPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
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
	if err = WriteInt8(w, p.Yaw); err != nil {
		return
	}
	if err = WriteInt8(w, p.Pitch); err != nil {
		return
	}
	return WriteBool(w, p.OnGround)
}

// --- RelativeEntityMovePacket (Clientbound) ---

type RelativeEntityMovePacket struct {
	EntityID   int32
	DX, DY, DZ int16
	OnGround   bool
}

func (p *RelativeEntityMovePacket) ID(version string) uint32                   { return 0x2F }
func (p *RelativeEntityMovePacket) State() State                               { return StatePlay }
func (p *RelativeEntityMovePacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *RelativeEntityMovePacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteInt16(w, p.DX); err != nil {
		return
	}
	if err = WriteInt16(w, p.DY); err != nil {
		return
	}
	if err = WriteInt16(w, p.DZ); err != nil {
		return
	}
	return WriteBool(w, p.OnGround)
}

// --- EntityMoveLookPacket (Clientbound) ---

type EntityMoveLookPacket struct {
	EntityID   int32
	DX, DY, DZ int16
	Yaw, Pitch int8
	OnGround   bool
}

func (p *EntityMoveLookPacket) ID(version string) uint32                   { return 0x30 }
func (p *EntityMoveLookPacket) State() State                               { return StatePlay }
func (p *EntityMoveLookPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityMoveLookPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteInt16(w, p.DX); err != nil {
		return
	}
	if err = WriteInt16(w, p.DY); err != nil {
		return
	}
	if err = WriteInt16(w, p.DZ); err != nil {
		return
	}
	if err = WriteInt8(w, p.Yaw); err != nil {
		return
	}
	if err = WriteInt8(w, p.Pitch); err != nil {
		return
	}
	return WriteBool(w, p.OnGround)
}

// --- EntityLookPacket (Clientbound) ---

type EntityLookPacket struct {
	EntityID   int32
	Yaw, Pitch int8
	OnGround   bool
}

func (p *EntityLookPacket) ID(version string) uint32                   { return 0x32 }
func (p *EntityLookPacket) State() State                               { return StatePlay }
func (p *EntityLookPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityLookPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteInt8(w, p.Yaw); err != nil {
		return
	}
	if err = WriteInt8(w, p.Pitch); err != nil {
		return
	}
	return WriteBool(w, p.OnGround)
}

// --- EntityDestroyPacket (Clientbound) ---
type EntityDestroyPacket struct {
	EntityIDs []int32
}

func (p *EntityDestroyPacket) ID(version string) uint32                   { return 0x44 }
func (p *EntityDestroyPacket) State() State                               { return StatePlay }
func (p *EntityDestroyPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityDestroyPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, int32(len(p.EntityIDs))); err != nil {
		return
	}
	for _, id := range p.EntityIDs {
		if err = WriteVarInt(w, id); err != nil {
			return
		}
	}
	return
}

func init() {
	// Clientbound
	RegisterPacket(StatePlay, Clientbound, 0x01, func() Packet { return &SpawnEntityPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x02, func() Packet { return &SpawnExperienceOrbPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x40, func() Packet { return &PlayerInfoUpdatePacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x3F, func() Packet { return &PlayerInfoRemovePacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x6C, func() Packet { return &EntityTeleportPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x2F, func() Packet { return &RelativeEntityMovePacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x30, func() Packet { return &EntityMoveLookPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x32, func() Packet { return &EntityLookPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x44, func() Packet { return &EntityDestroyPacket{} })

	// Serverbound packets for entities are more interaction-based (UseEntity, etc.)
	// and will be in a different file.
}
