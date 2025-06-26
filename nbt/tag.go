package nbt

import (
	"fmt"
	"io"
)

// TagID represents the type of an NBT tag.
type TagID = byte

const (
	TagEnd TagID = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

// TagTypeNames maps TagIDs to their string representations.
var TagTypeNames = map[TagID]string{
	TagEnd:       "TAG_End",
	TagByte:      "TAG_Byte",
	TagShort:     "TAG_Short",
	TagInt:       "TAG_Int",
	TagLong:      "TAG_Long",
	TagFloat:     "TAG_Float",
	TagDouble:    "TAG_Double",
	TagByteArray: "TAG_Byte_Array",
	TagString:    "TAG_String",
	TagList:      "TAG_List",
	TagCompound:  "TAG_Compound",
	TagIntArray:  "TAG_Int_Array",
	TagLongArray: "TAG_Long_Array",
}

// Tag represents a single NBT tag.
type Tag interface {
	// ID returns the numeric type ID of the tag.
	ID() TagID
	// String returns the SNBT representation of the tag.
	String() string
	// write writes the binary payload of the tag to the writer.
	write(w io.Writer) error
	// read reads the binary payload of the tag from the reader.
	read(r io.Reader) error
	// Copy creates a deep copy of the tag.
	Copy() Tag
}

// NamedTag represents a tag with a name, used as the root of an NBT file.
type NamedTag struct {
	Name string
	Tag  Tag
}

// newTag creates a new tag of the given type ID.
func newTag(id TagID) (Tag, error) {
	switch id {
	case TagEnd:
		return new(EndTag), nil
	case TagByte:
		return new(ByteTag), nil
	case TagShort:
		return new(ShortTag), nil
	case TagInt:
		return new(IntTag), nil
	case TagLong:
		return new(LongTag), nil
	case TagFloat:
		return new(FloatTag), nil
	case TagDouble:
		return new(DoubleTag), nil
	case TagByteArray:
		return new(ByteArrayTag), nil
	case TagString:
		return new(StringTag), nil
	case TagList:
		return new(ListTag), nil
	case TagCompound:
		return NewCompoundTag(), nil
	case TagIntArray:
		return new(IntArrayTag), nil
	case TagLongArray:
		return new(LongArrayTag), nil
	default:
		return nil, fmt.Errorf("invalid tag id: %d", id)
	}
}
