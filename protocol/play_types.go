package protocol

import (
    "github.com/Advik-B/Golem/nbt"
    "io"
)

// --- Slot Data (1.20.5+) ---

// Slot represents an item stack in an inventory using the modern component system.
type Slot struct {
    ItemCount             int32
    ItemID                int32
    AddedComponentCount   int32
    RemovedComponentCount int32
    Components            []SlotComponent
    RemoveComponents      []int32 // These are SlotComponentType enums
}

func (s *Slot) ReadFrom(r io.Reader) (err error) {
    if s.ItemCount, err = ReadVarInt(r); err != nil || s.ItemCount == 0 {
        return
    }
    if s.ItemID, err = ReadVarInt(r); err != nil {
        return
    }
    if s.AddedComponentCount, err = ReadVarInt(r); err != nil {
        return
    }
    if s.RemovedComponentCount, err = ReadVarInt(r); err != nil {
        return
    }

    s.Components = make([]SlotComponent, s.AddedComponentCount)
    for i := int32(0); i < s.AddedComponentCount; i++ {
        if err = s.Components[i].ReadFrom(r); err != nil {
            return
        }
    }

    s.RemoveComponents = make([]int32, s.RemovedComponentCount)
    for i := int32(0); i < s.RemovedComponentCount; i++ {
        if s.RemoveComponents[i], err = ReadVarInt(r); err != nil {
            return
        }
    }
    return
}

func (s *Slot) WriteTo(w io.Writer) (err error) {
    if err = WriteVarInt(w, s.ItemCount); err != nil {
        return
    }
    if s.ItemCount == 0 {
        return nil
    }
    if err = WriteVarInt(w, s.ItemID); err != nil {
        return
    }
    if err = WriteVarInt(w, int32(len(s.Components))); err != nil {
        return
    }
    if err = WriteVarInt(w, int32(len(s.RemoveComponents))); err != nil {
        return
    }
    for i := range s.Components {
        if err = s.Components[i].WriteTo(w); err != nil {
            return
        }
    }
    for _, compType := range s.RemoveComponents {
        if err = WriteVarInt(w, compType); err != nil {
            return
        }
    }
    return
}

// SlotComponent represents a single component of an item stack.
type SlotComponent struct {
    Type int32       // SlotComponentType enum
    Data interface{} // Holds a struct specific to the component type
}

func (sc *SlotComponent) ReadFrom(r io.Reader) error {
    var err error
    sc.Type, err = ReadVarInt(r)
    if err != nil {
        return err
    }
    // In a full implementation, a switch on sc.Type would call a specific
    // reader for sc.Data. For now, we assume it's read externally or is NBT.
    // As a placeholder, many simple types are just VarInts or bools.
    // For example, 'max_stack_size' is just a VarInt.
    // 'custom_data' is NBT.
    // We will treat all data as Anonymous NBT for now for simplicity.
    sc.Data, err = ReadNBT(r) // Simplified for now
    return err
}

func (sc *SlotComponent) WriteTo(w io.Writer) error {
    if err := WriteVarInt(w, sc.Type); err != nil {
        return err
    }
    // Simplified for now
    return WriteNBT(w, sc.Data.(nbt.Tag))
}

// --- Spawn Info ---
// Used in Login and Respawn packets.
type SpawnInfo struct {
    Dimension        int32
    Name             string
    HashedSeed       int64
    GameMode         int8
    PreviousGameMode uint8
    IsDebug          bool
    IsFlat           bool
    DeathLocation    *struct {
        DimensionName string
        Location      int64 // Packed position
    }
    PortalCooldown int32
    SeaLevel       int32
}

func (si *SpawnInfo) ReadFrom(r io.Reader) (err error) {
    if si.Dimension, err = ReadVarInt(r); err != nil {
        return
    }
    if si.Name, err = ReadString(r); err != nil {
        return
    }
    if si.HashedSeed, err = ReadInt64(r); err != nil {
        return
    }
    if si.GameMode, err = ReadInt8(r); err != nil {
        return
    }
    if si.PreviousGameMode, err = ReadUint8(r); err != nil {
        return
    }
    if si.IsDebug, err = ReadBool(r); err != nil {
        return
    }
    if si.IsFlat, err = ReadBool(r); err != nil {
        return
    }

    hasDeathLoc, err := ReadBool(r)
    if err != nil {
        return
    }
    if hasDeathLoc {
        si.DeathLocation = &struct {
            DimensionName string
            Location      int64
        }{}
        if si.DeathLocation.DimensionName, err = ReadString(r); err != nil {
            return
        }
        if si.DeathLocation.Location, err = ReadInt64(r); err != nil {
            return
        }
    }

    if si.PortalCooldown, err = ReadVarInt(r); err != nil {
        return
    }
    si.SeaLevel, err = ReadVarInt(r)
    return
}

func (si *SpawnInfo) WriteTo(w io.Writer) (err error) {
    if err = WriteVarInt(w, si.Dimension); err != nil {
        return
    }
    if err = WriteString(w, si.Name); err != nil {
        return
    }
    if err = WriteInt64(w, si.HashedSeed); err != nil {
        return
    }
    if err = WriteInt8(w, si.GameMode); err != nil {
        return
    }
    if err = WriteUint8(w, si.PreviousGameMode); err != nil {
        return
    }
    if err = WriteBool(w, si.IsDebug); err != nil {
        return
    }
    if err = WriteBool(w, si.IsFlat); err != nil {
        return
    }

    hasDeathLoc := si.DeathLocation != nil
    if err = WriteBool(w, hasDeathLoc); err != nil {
        return
    }
    if hasDeathLoc {
        if err = WriteString(w, si.DeathLocation.DimensionName); err != nil {
            return
        }
        if err = WriteInt64(w, si.DeathLocation.Location); err != nil {
            return
        }
    }

    if err = WriteVarInt(w, si.PortalCooldown); err != nil {
        return
    }
    return WriteVarInt(w, si.SeaLevel)
}
