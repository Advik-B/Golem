package protocol

import (
	"github.com/Advik-B/Golem/protocol/codec"
	"github.com/google/uuid"
)

// ResourceLocation represents a namespaced ID, like "minecraft:stone".
type ResourceLocation string

func (rl *ResourceLocation) ReadFrom(r *codec.PacketBuffer) error {
	s, err := r.ReadString(32767)
	*rl = ResourceLocation(s)
	return err
}

func (rl *ResourceLocation) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteString(string(*rl))
}

// BlockPos represents the 3D integer coordinates of a block.
type BlockPos struct {
	X, Y, Z int
}

func (bp *BlockPos) ReadFrom(r *codec.PacketBuffer) error {
	val, err := r.ReadLong()
	if err != nil {
		return err
	}
	bp.X = int(val >> 38)
	bp.Y = int(val & 0xFFF)
	bp.Z = int(val << 26 >> 38)
	return nil
}

func (bp *BlockPos) WriteTo(w *codec.PacketBuffer) error {
	val := (int64(bp.X&0x3FFFFFF) << 38) | (int64(bp.Z&0x3FFFFFF) << 12) | int64(bp.Y&0xFFF)
	return w.WriteLong(val)
}

// Component represents a JSON chat component.
// For now, we'll treat it as a raw string.
type Component string

func (c *Component) ReadFrom(r *codec.PacketBuffer) error {
	s, err := r.ReadString(262144)
	*c = Component(s)
	return err
}

func (c *Component) WriteTo(w *codec.PacketBuffer) error {
	return w.WriteString(string(*c))
}

// GameProfile represents a player's profile, including UUID, name, and properties.
type GameProfile struct {
	ID   uuid.UUID
	Name string
	// Properties would be a more complex type in a full implementation
}

func (gp *GameProfile) ReadFrom(r *codec.PacketBuffer) error {
	var err error
	if gp.ID, err = r.ReadUUID(); err != nil {
		return err
	}
	gp.Name, err = r.ReadString(16)
	return err
}

func (gp *GameProfile) WriteTo(w *codec.PacketBuffer) error {
	w.WriteUUID(gp.ID)
	w.WriteString(gp.Name)
	// Write properties (0 for now)
	w.WriteVarInt(0)
	return nil
}
