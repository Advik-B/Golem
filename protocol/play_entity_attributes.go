// File: protocol/play_entity_attributes.go
package protocol

import (
	"fmt"
	"io"

	"github.com/Advik-B/Golem/nbt"
	"github.com/google/uuid"
)

// --- EntityMetadataPacket (Clientbound) ---
type EntityMetadataPacket struct {
	EntityID int32
	Metadata []MetadataEntry
}

func (p *EntityMetadataPacket) ID(version string) uint32                   { return 0x56 }
func (p *EntityMetadataPacket) State() State                               { return StatePlay }
func (p *EntityMetadataPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityMetadataPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}

	for _, entry := range p.Metadata {
		if err = WriteUint8(w, entry.Key); err != nil {
			return
		}
		if err = WriteVarInt(w, int32(entry.Type)); err != nil {
			return
		}

		switch entry.Type {
		case MetadataByte:
			err = WriteInt8(w, entry.Value.(int8))
		case MetadataInt:
			err = WriteVarInt(w, entry.Value.(int32))
		case MetadataLong:
			err = WriteVarLong(w, entry.Value.(int64))
		case MetadataFloat:
			err = WriteFloat32(w, entry.Value.(float32))
		case MetadataString:
			err = WriteString(w, entry.Value.(string))
		case MetadataComponent:
			err = WriteNBT(w, entry.Value.(nbt.Tag))
		case MetadataOptionalComponent:
			hasValue := entry.Value != nil
			if err = WriteBool(w, hasValue); err != nil {
				break
			}
			if hasValue {
				err = WriteNBT(w, entry.Value.(nbt.Tag))
			}
		case MetadataItemStack:
			err = entry.Value.(*Slot).WriteTo(w) // Assuming value is a pointer to a Slot
		case MetadataBoolean:
			err = WriteBool(w, entry.Value.(bool))
		case MetadataRotations:
			rot := entry.Value.([3]float32)
			if err = WriteFloat32(w, rot[0]); err != nil {
				break
			}
			if err = WriteFloat32(w, rot[1]); err != nil {
				break
			}
			err = WriteFloat32(w, rot[2])
		case MetadataBlockPos:
			pos := entry.Value.([3]int)
			err = WritePosition(w, pos[0], pos[1], pos[2])
		case MetadataOptionalBlockPos:
			hasValue := entry.Value != nil
			if err = WriteBool(w, hasValue); err != nil {
				break
			}
			if hasValue {
				pos := entry.Value.([3]int)
				err = WritePosition(w, pos[0], pos[1], pos[2])
			}
		case MetadataDirection:
			err = WriteVarInt(w, entry.Value.(int32))
		case MetadataOptionalUUID:
			hasValue := entry.Value != nil
			if err = WriteBool(w, hasValue); err != nil {
				break
			}
			if hasValue {
				err = WriteUUID(w, entry.Value.(uuid.UUID))
			}
		case MetadataBlockState:
			err = WriteVarInt(w, entry.Value.(int32))
		case MetadataOptionalBlockState:
			err = WriteVarInt(w, entry.Value.(int32)) // 0 for none, value+1 for present
		case MetadataCompoundTag:
			err = WriteNBT(w, entry.Value.(nbt.Tag))
		// Other types would follow here...
		default:
			return fmt.Errorf("unhandled metadata type %d for writing", entry.Type)
		}

		if err != nil {
			return fmt.Errorf("error writing metadata type %d: %w", entry.Type, err)
		}
	}

	return WriteUint8(w, 0xFF) // End of metadata
}

// --- EntityVelocityPacket (Clientbound) ---
type EntityVelocityPacket struct {
	EntityID         int32
	VelX, VelY, VelZ int16
}

func (p *EntityVelocityPacket) ID(version string) uint32                   { return 0x58 }
func (p *EntityVelocityPacket) State() State                               { return StatePlay }
func (p *EntityVelocityPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityVelocityPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
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

// --- EntityEquipmentPacket (Clientbound) ---
type EntityEquipmentPacket struct {
	EntityID   int32
	Equipments []struct {
		Slot int8
		Item Slot
	}
}

func (p *EntityEquipmentPacket) ID(version string) uint32                   { return 0x59 }
func (p *EntityEquipmentPacket) State() State                               { return StatePlay }
func (p *EntityEquipmentPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityEquipmentPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	// Implements topBitSetTerminatedArray logic
	for i, equip := range p.Equipments {
		slotValue := equip.Slot
		if i < len(p.Equipments)-1 {
			// FIX: Cast to uint8 to perform the bitwise OR, then cast back to int8.
			slotValue = int8(uint8(slotValue) | 0x80)
		}
		if err = WriteInt8(w, slotValue); err != nil {
			return
		}
		if err = equip.Item.WriteTo(w); err != nil {
			return
		}
	}
	return
}

// --- EntityEffectPacket (Clientbound) ---
type EntityEffectPacket struct {
	EntityID  int32
	EffectID  int32
	Amplifier int32
	Duration  int32
	Flags     uint8
}

func (p *EntityEffectPacket) ID(version string) uint32                   { return 0x72 }
func (p *EntityEffectPacket) State() State                               { return StatePlay }
func (p *EntityEffectPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityEffectPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteVarInt(w, p.EffectID); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Amplifier); err != nil {
		return
	}
	if err = WriteVarInt(w, p.Duration); err != nil {
		return
	}
	return WriteUint8(w, p.Flags)
}

// --- RemoveEntityEffectPacket (Clientbound) ---
type RemoveEntityEffectPacket struct {
	EntityID int32
	EffectID int32
}

func (p *RemoveEntityEffectPacket) ID(version string) uint32                   { return 0x45 }
func (p *RemoveEntityEffectPacket) State() State                               { return StatePlay }
func (p *RemoveEntityEffectPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *RemoveEntityEffectPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	return WriteVarInt(w, p.EffectID)
}

// --- EntityHeadRotationPacket (Clientbound) ---
type EntityHeadRotationPacket struct {
	EntityID int32
	HeadYaw  int8
}

func (p *EntityHeadRotationPacket) ID(version string) uint32                   { return 0x47 }
func (p *EntityHeadRotationPacket) State() State                               { return StatePlay }
func (p *EntityHeadRotationPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *EntityHeadRotationPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	return WriteInt8(w, p.HeadYaw)
}

// --- AttachEntityPacket (Clientbound) ---
type AttachEntityPacket struct {
	EntityID  int32
	VehicleID int32
}

func (p *AttachEntityPacket) ID(version string) uint32                   { return 0x57 }
func (p *AttachEntityPacket) State() State                               { return StatePlay }
func (p *AttachEntityPacket) ReadFrom(r io.Reader, version string) error { return nil }
func (p *AttachEntityPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteInt32(w, p.EntityID); err != nil {
		return
	}
	return WriteInt32(w, p.VehicleID)
}

// --- SetPassengersPacket (Clientbound) ---
type SetPassengersPacket struct {
	EntityID   int32
	Passengers []int32
}

func (p *SetPassengersPacket) ID(version string) uint32                   { return 0x5C }
func (p *SetPassengersPacket) State() State                               { return StatePlay }
func (p *SetPassengersPacket) ReadFrom(r io.Reader, version string) error { return nil } // Server only
func (p *SetPassengersPacket) WriteTo(w io.Writer, version string) (err error) {
	if err = WriteVarInt(w, p.EntityID); err != nil {
		return
	}
	if err = WriteVarInt(w, int32(len(p.Passengers))); err != nil {
		return
	}
	for _, passengerID := range p.Passengers {
		if err = WriteVarInt(w, passengerID); err != nil {
			return
		}
	}
	return
}

func init() {
	// Clientbound
	RegisterPacket(StatePlay, Clientbound, 0x56, func() Packet { return &EntityMetadataPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x58, func() Packet { return &EntityVelocityPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x59, func() Packet { return &EntityEquipmentPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x72, func() Packet { return &EntityEffectPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x45, func() Packet { return &RemoveEntityEffectPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x47, func() Packet { return &EntityHeadRotationPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x57, func() Packet { return &AttachEntityPacket{} })
	RegisterPacket(StatePlay, Clientbound, 0x5C, func() Packet { return &SetPassengersPacket{} })
}
