package nbt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strings"
)

// --- Primitive Tags ---

type EndTag struct{}

func (t *EndTag) ID() TagID               { return TagEnd }
func (t *EndTag) String() string          { return "" }
func (t *EndTag) write(w io.Writer) error { return nil }
func (t *EndTag) read(r io.Reader) error  { return nil }
func (t *EndTag) Copy() Tag               { return new(EndTag) }

type ByteTag struct{ Value byte }

func (t *ByteTag) ID() TagID               { return TagByte }
func (t *ByteTag) String() string          { return fmt.Sprintf("%db", t.Value) }
func (t *ByteTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *ByteTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *ByteTag) Copy() Tag               { return &ByteTag{Value: t.Value} }

type ShortTag struct{ Value int16 }

func (t *ShortTag) ID() TagID               { return TagShort }
func (t *ShortTag) String() string          { return fmt.Sprintf("%ds", t.Value) }
func (t *ShortTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *ShortTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *ShortTag) Copy() Tag               { return &ShortTag{Value: t.Value} }

type IntTag struct{ Value int32 }

func (t *IntTag) ID() TagID               { return TagInt }
func (t *IntTag) String() string          { return fmt.Sprintf("%d", t.Value) }
func (t *IntTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *IntTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *IntTag) Copy() Tag               { return &IntTag{Value: t.Value} }

type LongTag struct{ Value int64 }

func (t *LongTag) ID() TagID               { return TagLong }
func (t *LongTag) String() string          { return fmt.Sprintf("%dL", t.Value) }
func (t *LongTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *LongTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *LongTag) Copy() Tag               { return &LongTag{Value: t.Value} }

type FloatTag struct{ Value float32 }

func (t *FloatTag) ID() TagID               { return TagFloat }
func (t *FloatTag) String() string          { return fmt.Sprintf("%gf", t.Value) }
func (t *FloatTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *FloatTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *FloatTag) Copy() Tag               { return &FloatTag{Value: t.Value} }

type DoubleTag struct{ Value float64 }

func (t *DoubleTag) ID() TagID               { return TagDouble }
func (t *DoubleTag) String() string          { return fmt.Sprintf("%gd", t.Value) }
func (t *DoubleTag) write(w io.Writer) error { return binary.Write(w, binary.BigEndian, t.Value) }
func (t *DoubleTag) read(r io.Reader) error  { return binary.Read(r, binary.BigEndian, &t.Value) }
func (t *DoubleTag) Copy() Tag               { return &DoubleTag{Value: t.Value} }

type StringTag struct{ Value string }

func (t *StringTag) ID() TagID               { return TagString }
func (t *StringTag) String() string          { return quoteAndEscape(t.Value) }
func (t *StringTag) write(w io.Writer) error { return writeUTF(w, t.Value) }
func (t *StringTag) read(r io.Reader) error {
	var err error
	t.Value, err = readUTF(r)
	return err
}
func (t *StringTag) Copy() Tag { return &StringTag{Value: t.Value} }

// --- Array Tags ---

type ByteArrayTag struct{ Value []byte }

func (t *ByteArrayTag) ID() TagID { return TagByteArray }
func (t *ByteArrayTag) String() string {
	var sb strings.Builder
	sb.WriteString("[B;")
	for i, b := range t.Value {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", b))
	}
	sb.WriteString("]")
	return sb.String()
}
func (t *ByteArrayTag) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(t.Value))); err != nil {
		return err
	}
	_, err := w.Write(t.Value)
	return err
}
func (t *ByteArrayTag) read(r io.Reader) error {
	var size int32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("negative array size: %d", size)
	}
	t.Value = make([]byte, size)
	_, err := io.ReadFull(r, t.Value)
	return err
}
func (t *ByteArrayTag) Copy() Tag {
	c := make([]byte, len(t.Value))
	copy(c, t.Value)
	return &ByteArrayTag{Value: c}
}

type IntArrayTag struct{ Value []int32 }

func (t *IntArrayTag) ID() TagID { return TagIntArray }
func (t *IntArrayTag) String() string {
	var sb strings.Builder
	sb.WriteString("[I;")
	for i, v := range t.Value {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", v))
	}
	sb.WriteString("]")
	return sb.String()
}
func (t *IntArrayTag) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(t.Value))); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, t.Value)
}
func (t *IntArrayTag) read(r io.Reader) error {
	var size int32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("negative array size: %d", size)
	}
	t.Value = make([]int32, size)
	return binary.Read(r, binary.BigEndian, &t.Value)
}
func (t *IntArrayTag) Copy() Tag {
	c := make([]int32, len(t.Value))
	copy(c, t.Value)
	return &IntArrayTag{Value: c}
}

type LongArrayTag struct{ Value []int64 }

func (t *LongArrayTag) ID() TagID { return TagLongArray }
func (t *LongArrayTag) String() string {
	var sb strings.Builder
	sb.WriteString("[L;")
	for i, v := range t.Value {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%d", v))
	}
	sb.WriteString("]")
	return sb.String()
}
func (t *LongArrayTag) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, int32(len(t.Value))); err != nil {
		return err
	}
	return binary.Write(w, binary.BigEndian, t.Value)
}
func (t *LongArrayTag) read(r io.Reader) error {
	var size int32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("negative array size: %d", size)
	}
	t.Value = make([]int64, size)
	return binary.Read(r, binary.BigEndian, &t.Value)
}
func (t *LongArrayTag) Copy() Tag {
	c := make([]int64, len(t.Value))
	copy(c, t.Value)
	return &LongArrayTag{Value: c}
}

// --- List Tag ---

type ListTag struct {
	Type  TagID
	Value []Tag
}

func (t *ListTag) ID() TagID { return TagList }
func (t *ListTag) String() string {
	if len(t.Value) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i, tag := range t.Value {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(tag.String())
	}
	sb.WriteString("]")
	return sb.String()
}
func (t *ListTag) write(w io.Writer) error {
	if err := binary.Write(w, binary.BigEndian, t.Type); err != nil {
		return err
	}
	if err := binary.Write(w, binary.BigEndian, int32(len(t.Value))); err != nil {
		return err
	}
	for _, tag := range t.Value {
		if err := tag.write(w); err != nil {
			return err
		}
	}
	return nil
}
func (t *ListTag) read(r io.Reader) error {
	if err := binary.Read(r, binary.BigEndian, &t.Type); err != nil {
		return err
	}
	var size int32
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("negative list size: %d", size)
	}
	t.Value = make([]Tag, size)
	for i := range t.Value {
		tag, err := newTag(t.Type)
		if err != nil {
			return err
		}
		if err := tag.read(r); err != nil {
			return err
		}
		t.Value[i] = tag
	}
	return nil
}
func (t *ListTag) Copy() Tag {
	c := &ListTag{Type: t.Type, Value: make([]Tag, len(t.Value))}
	for i, tag := range t.Value {
		c.Value[i] = tag.Copy()
	}
	return c
}
func (t *ListTag) Add(tag Tag) error {
	if len(t.Value) == 0 {
		t.Type = tag.ID()
	} else if t.Type != tag.ID() {
		return fmt.Errorf("cannot add tag of type %s to list of type %s", TagTypeNames[tag.ID()], TagTypeNames[t.Type])
	}
	t.Value = append(t.Value, tag)
	return nil
}

// --- Compound Tag ---

type CompoundTag struct {
	Value map[string]Tag
}

func NewCompoundTag() *CompoundTag {
	return &CompoundTag{Value: make(map[string]Tag)}
}
func (t *CompoundTag) ID() TagID { return TagCompound }
func (t *CompoundTag) String() string {
	if len(t.Value) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{")
	keys := make([]string, 0, len(t.Value))
	for k := range t.Value {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		if i > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(quoteAndEscape(k))
		sb.WriteString(":")
		sb.WriteString(t.Value[k].String())
	}
	sb.WriteString("}")
	return sb.String()
}
func (t *CompoundTag) write(w io.Writer) error {
	for name, tag := range t.Value {
		if err := binary.Write(w, binary.BigEndian, tag.ID()); err != nil {
			return err
		}
		if err := writeUTF(w, name); err != nil {
			return err
		}
		if err := tag.write(w); err != nil {
			return err
		}
	}
	return binary.Write(w, binary.BigEndian, TagEnd)
}
func (t *CompoundTag) read(r io.Reader) error {
	t.Value = make(map[string]Tag)
	for {
		var id TagID
		if err := binary.Read(r, binary.BigEndian, &id); err != nil {
			return err
		}
		if id == TagEnd {
			break
		}
		name, err := readUTF(r)
		if err != nil {
			return err
		}
		tag, err := newTag(id)
		if err != nil {
			return err
		}
		if err := tag.read(r); err != nil {
			return err
		}
		t.Value[name] = tag
	}
	return nil
}
func (t *CompoundTag) Copy() Tag {
	c := NewCompoundTag()
	for k, v := range t.Value {
		c.Value[k] = v.Copy()
	}
	return c
}
func (t *CompoundTag) Get(key string) (Tag, bool) {
	tag, ok := t.Value[key]
	return tag, ok
}
func (t *CompoundTag) Put(key string, tag Tag) {
	t.Value[key] = tag
}
func (t *CompoundTag) GetString(key string) (string, bool) {
	if tag, ok := t.Value[key]; ok {
		if st, ok := tag.(*StringTag); ok {
			return st.Value, true
		}
	}
	return "", false
}
func (t *CompoundTag) GetInt(key string) (int32, bool) {
	if tag, ok := t.Value[key]; ok {
		if it, ok := tag.(*IntTag); ok {
			return it.Value, true
		}
	}
	return 0, false
}
func (t *CompoundTag) GetCompound(key string) (*CompoundTag, bool) {
	if tag, ok := t.Value[key]; ok {
		if ct, ok := tag.(*CompoundTag); ok {
			return ct, true
		}
	}
	return nil, false
}
func (t *CompoundTag) GetList(key string) (*ListTag, bool) {
	if tag, ok := t.Value[key]; ok {
		if lt, ok := tag.(*ListTag); ok {
			return lt, true
		}
	}
	return nil, false
}

// --- UTF Helper ---

func readUTF(r io.Reader) (string, error) {
	var length uint16
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	// This decodes Java's modified UTF-8
	return decodeMUTF8(buf), nil
}

func writeUTF(w io.Writer, s string) error {
	// This encodes to Java's modified UTF-8
	encoded := encodeMUTF8(s)
	length := uint16(len(encoded))
	if err := binary.Write(w, binary.BigEndian, length); err != nil {
		return err
	}
	_, err := w.Write(encoded)
	return err
}

// MUTF8 encoding/decoding, simplified for this context.
// A full implementation handles surrogate pairs correctly.
func decodeMUTF8(b []byte) string {
	var buf bytes.Buffer
	for i := 0; i < len(b); {
		c := b[i]
		i++
		switch c >> 4 {
		case 0, 1, 2, 3, 4, 5, 6, 7:
			buf.WriteByte(c)
		case 12, 13:
			c2 := b[i]
			i++
			buf.WriteRune(rune(c&0x1F)<<6 | rune(c2&0x3F))
		case 14:
			c2 := b[i]
			i++
			c3 := b[i]
			i++
			buf.WriteRune(rune(c&0x0F)<<12 | rune(c2&0x3F)<<6 | rune(c3&0x3F))
		}
	}
	return buf.String()
}

func encodeMUTF8(s string) []byte {
	var buf bytes.Buffer
	for _, r := range s {
		if r >= 0x0001 && r <= 0x007F {
			buf.WriteByte(byte(r))
		} else if r > 0x07FF {
			buf.WriteByte(byte(0xE0 | (r >> 12 & 0x0F)))
			buf.WriteByte(byte(0x80 | (r >> 6 & 0x3F)))
			buf.WriteByte(byte(0x80 | (r & 0x3F)))
		} else { // also handles r == 0
			buf.WriteByte(byte(0xC0 | (r >> 6 & 0x1F)))
			buf.WriteByte(byte(0x80 | (r & 0x3F)))
		}
	}
	return buf.Bytes()
}

// --- String escaping for SNBT ---
var simpleValuePattern = `[A-Za-z0-9._+-]+` // Simplified from regex for performance
func isSimple(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '.' || r == '_' || r == '+' || r == '-') {
			return false
		}
	}
	return true
}

func quoteAndEscape(s string) string {
	var sb strings.Builder
	quote := `"`
	if strings.ContainsRune(s, '"') && !strings.ContainsRune(s, '\'') {
		quote = `'`
	}
	sb.WriteString(quote)
	for _, r := range s {
		if r == '\\' || r == rune(quote[0]) {
			sb.WriteRune('\\')
		}
		sb.WriteRune(r)
	}
	sb.WriteString(quote)
	return sb.String()
}
