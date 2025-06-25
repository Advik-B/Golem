package nbt

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

// CompareTags compares two NBT tags for equality. If partial is true, it checks if `a` is a subset of `b`.
// This version is more explicit and avoids reflect.DeepEqual pitfalls.
func CompareTags(a, b Tag, partial bool) bool {
	if a == b {
		return true
	}
	if a == nil && partial {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.ID() != b.ID() {
		return false
	}

	switch ta := a.(type) {
	case *EndTag:
		_, ok := b.(*EndTag)
		return ok
	case *ByteTag:
		tb, ok := b.(*ByteTag)
		return ok && ta.Value == tb.Value
	case *ShortTag:
		tb, ok := b.(*ShortTag)
		return ok && ta.Value == tb.Value
	case *IntTag:
		tb, ok := b.(*IntTag)
		return ok && ta.Value == tb.Value
	case *LongTag:
		tb, ok := b.(*LongTag)
		return ok && ta.Value == tb.Value
	case *FloatTag:
		tb, ok := b.(*FloatTag)
		return ok && ta.Value == tb.Value
	case *DoubleTag:
		tb, ok := b.(*DoubleTag)
		return ok && ta.Value == tb.Value
	case *StringTag:
		tb, ok := b.(*StringTag)
		return ok && ta.Value == tb.Value
	case *ByteArrayTag:
		tb, ok := b.(*ByteArrayTag)
		return ok && bytes.Equal(ta.Value, tb.Value)
	case *IntArrayTag:
		tb, ok := b.(*IntArrayTag)
		if !ok || len(ta.Value) != len(tb.Value) {
			return false
		}
		for i := range ta.Value {
			if ta.Value[i] != tb.Value[i] {
				return false
			}
		}
		return true
	case *LongArrayTag:
		tb, ok := b.(*LongArrayTag)
		if !ok || len(ta.Value) != len(tb.Value) {
			return false
		}
		for i := range ta.Value {
			if ta.Value[i] != tb.Value[i] {
				return false
			}
		}
		return true
	case *ListTag:
		tb, ok := b.(*ListTag)
		if !ok || ta.Type != tb.Type {
			return false
		}
		if !partial && len(ta.Value) != len(tb.Value) {
			return false
		}
		if partial {
		outer:
			for _, valA := range ta.Value {
				for _, valB := range tb.Value {
					if CompareTags(valA, valB, true) {
						continue outer
					}
				}
				return false // No match found for valA
			}
			return true
		}
		// Full comparison
		if len(ta.Value) != len(tb.Value) {
			return false
		}
		for i := range ta.Value {
			if !CompareTags(ta.Value[i], tb.Value[i], false) {
				return false
			}
		}
		return true
	case *CompoundTag:
		tb, ok := b.(*CompoundTag)
		if !ok {
			return false
		}
		if !partial && len(ta.Value) != len(tb.Value) {
			return false
		}
		for key, valA := range ta.Value {
			valB, ok := tb.Value[key]
			if !ok || !CompareTags(valA, valB, partial) {
				return false
			}
		}
		return true
	default:
		// Should not be reached if all types are handled
		return false
	}
}

// --- Structure Template Packing/Unpacking ---

// StructureToSnbt converts a structure CompoundTag to its SNBT representation,
// applying the special packing rules from NbtUtils.
func StructureToSnbt(tag *CompoundTag) (string, error) {
	packed, err := packStructureTemplate(tag.Copy().(*CompoundTag))
	if err != nil {
		return "", err
	}
	return ToSNBT(packed), nil
}

// SnbtToStructure converts an SNBT string back into a structure CompoundTag,
// applying the special unpacking rules.
func SnbtToStructure(snbt string) (*CompoundTag, error) {
	parsed, err := ParseSNBT(snbt)
	if err != nil {
		return nil, err
	}
	return unpackStructureTemplate(parsed)
}

func packStructureTemplate(tag *CompoundTag) (*CompoundTag, error) {
	// Pack palette
	paletteList, _ := tag.GetList("palette")
	if palettesTag, ok := tag.Get("palettes"); ok {
		if palettes, ok := palettesTag.(*ListTag); ok && len(palettes.Value) > 0 {
			if p, ok := palettes.Value[0].(*ListTag); ok {
				paletteList = p
			}
		}
	}

	packedPalette := &ListTag{Type: TagString}
	for _, pTag := range paletteList.Value {
		if cTag, ok := pTag.(*CompoundTag); ok {
			packedPalette.Add(&StringTag{Value: packBlockState(cTag)})
		}
	}
	tag.Put("palette", packedPalette)

	// Pack blocks
	blocks, ok := tag.GetList("blocks")
	if !ok {
		return tag, nil // No blocks to pack
	}

	dataList := &ListTag{Type: TagCompound}
	for _, blockTag := range blocks.Value {
		block, _ := blockTag.(*CompoundTag)
		stateIdx, _ := block.GetInt("state")
		packedBlock := block.Copy().(*CompoundTag)
		if int(stateIdx) < len(packedPalette.Value) {
			packedState, _ := packedPalette.Value[int(stateIdx)].(*StringTag)
			packedBlock.Put("state", &StringTag{Value: packedState.Value})
		} else {
			return nil, fmt.Errorf("state index %d out of bounds for palette size %d", stateIdx, len(packedPalette.Value))
		}
		dataList.Add(packedBlock)
	}
	tag.Put("data", dataList)
	delete(tag.Value, "blocks")

	return tag, nil
}

func unpackStructureTemplate(tag *CompoundTag) (*CompoundTag, error) {
	palette, ok := tag.GetList("palette")
	if !ok {
		return nil, fmt.Errorf("structure missing palette")
	}

	unpackedPalette := &ListTag{Type: TagCompound}
	stateMap := make(map[string]int32)
	for i, packedStateTag := range palette.Value {
		packedState, _ := packedStateTag.(*StringTag)
		unpacked := unpackBlockState(packedState.Value)
		unpackedPalette.Add(unpacked)
		stateMap[packedState.Value] = int32(i)
	}
	tag.Put("palette", unpackedPalette)

	data, ok := tag.GetList("data")
	if !ok {
		return tag, nil // No data to unpack
	}

	blocksList := &ListTag{Type: TagCompound}
	for _, packedBlockTag := range data.Value {
		packedBlock, _ := packedBlockTag.(*CompoundTag)
		stateStr, _ := packedBlock.GetString("state")

		stateIdx, ok := stateMap[stateStr]
		if !ok {
			return nil, fmt.Errorf("state '%s' not found in palette", stateStr)
		}
		unpackedBlock := packedBlock.Copy().(*CompoundTag)
		unpackedBlock.Put("state", &IntTag{Value: stateIdx})
		blocksList.Add(unpackedBlock)
	}
	tag.Put("blocks", blocksList)
	delete(tag.Value, "data")

	return tag, nil
}

func packBlockState(tag *CompoundTag) string {
	name, _ := tag.GetString("Name")
	var sb strings.Builder
	sb.WriteString(name)
	props, ok := tag.GetCompound("Properties")
	if !ok || len(props.Value) == 0 {
		return sb.String()
	}

	var keys []string
	for k := range props.Value {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sb.WriteString("{")
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		v, _ := props.GetString(k)
		sb.WriteString(fmt.Sprintf("%s:%s", k, v))
	}
	sb.WriteString("}")
	return sb.String()
}

func unpackBlockState(s string) *CompoundTag {
	tag := NewCompoundTag()
	bracket := strings.Index(s, "{")
	if bracket == -1 {
		tag.Put("Name", &StringTag{Value: s})
		return tag
	}

	name := s[:bracket]
	tag.Put("Name", &StringTag{Value: name})

	propsStr := s[bracket+1 : strings.LastIndex(s, "}")]
	props := NewCompoundTag()
	pairs := strings.Split(propsStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 {
			props.Put(parts[0], &StringTag{Value: parts[1]})
		}
	}
	tag.Put("Properties", props)
	return tag
}
