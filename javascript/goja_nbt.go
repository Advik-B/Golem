package javascript

import (
	"bytes"
	"fmt"
	"github.com/Advik-B/Golem/nbt"
	"github.com/dop251/goja"
)

// nativeTag is a hidden property used to store the original Go nbt.Tag on a goja.Object proxy.
const nativeTag = "__native_nbt_tag__"

// NbtModule is a wrapper for the NBT library that can be injected into a goja.Runtime.
type NbtModule struct {
	runtime *goja.Runtime
}

// NewNbtModule creates the bindings for the nbt library.
func NewNbtModule(runtime *goja.Runtime) *goja.Object {
	m := &NbtModule{runtime: runtime}

	obj := runtime.NewObject()
	obj.Set("parse", m.parseSNBT)
	obj.Set("toPretty", m.toPrettySNBT)
	obj.Set("toCompact", m.toCompactSNBT)
	obj.Set("read", func(buffer goja.Value) *goja.Object { return m.readBinary(buffer, false) })
	obj.Set("readCompressed", func(buffer goja.Value) *goja.Object { return m.readBinary(buffer, true) })
	obj.Set("compare", m.compare)

	// Factory functions for creating new tags
	obj.Set("newByte", func(v int8) *goja.Object { return m.toProxy(&nbt.ByteTag{Value: v}) })
	obj.Set("newShort", func(v int16) *goja.Object { return m.toProxy(&nbt.ShortTag{Value: v}) })
	obj.Set("newInt", func(v int32) *goja.Object { return m.toProxy(&nbt.IntTag{Value: v}) })
	obj.Set("newLong", func(v int64) *goja.Object { return m.toProxy(&nbt.LongTag{Value: v}) })
	obj.Set("newFloat", func(v float32) *goja.Object { return m.toProxy(&nbt.FloatTag{Value: v}) })
	obj.Set("newDouble", func(v float64) *goja.Object { return m.toProxy(&nbt.DoubleTag{Value: v}) })
	obj.Set("newString", func(v string) *goja.Object { return m.toProxy(&nbt.StringTag{Value: v}) })
	obj.Set("newByteArray", func(v []int8) *goja.Object { return m.toProxy(&nbt.ByteArrayTag{Value: v}) })
	obj.Set("newIntArray", func(v []int32) *goja.Object { return m.toProxy(&nbt.IntArrayTag{Value: v}) })
	obj.Set("newLongArray", func(v []int64) *goja.Object { return m.toProxy(&nbt.LongArrayTag{Value: v}) })
	obj.Set("newList", func() *goja.Object { return m.toProxy(&nbt.ListTag{}) })
	obj.Set("newCompound", func() *goja.Object { return m.toProxy(nbt.NewCompoundTag()) })

	return obj
}

// toProxy converts a Go nbt.Tag into its JavaScript proxy object.
func (m *NbtModule) toProxy(tag nbt.Tag) *goja.Object {
	if tag == nil {
		return nil
	}
	obj := m.runtime.NewObject()
	obj.Set(nativeTag, tag)

	obj.Set("toString", func() string { return nbt.ToCompactSNBT(tag) })
	obj.Set("toPretty", func() string { return nbt.ToPrettySNBT(tag) })
	obj.Set("id", func() nbt.TagID { return tag.ID() })
	obj.Set("typeName", func() string { return nbt.TagTypeNames[tag.ID()] })
	obj.Set("copy", func() *goja.Object { return m.toProxy(tag.Copy()) })
	obj.Set("write", func(name ...string) goja.Value {
		rootName := ""
		if len(name) > 0 {
			rootName = name[0]
		}
		return m.writeBinary(tag, rootName, false)
	})
	obj.Set("writeCompressed", func(name ...string) goja.Value {
		rootName := ""
		if len(name) > 0 {
			rootName = name[0]
		}
		return m.writeBinary(tag, rootName, true)
	})

	switch t := tag.(type) {
	case *nbt.CompoundTag:
		obj.Set("get", func(key string) *goja.Object {
			val, _ := t.Get(key)
			return m.toProxy(val)
		})
		obj.Set("set", func(key string, val *goja.Object) {
			t.Put(key, m.fromProxy(val))
		})
		obj.Set("keys", func() []string {
			keys := make([]string, 0, len(t.Value))
			for k := range t.Value {
				keys = append(keys, k)
			}
			return keys
		})
	case *nbt.ListTag:
		obj.Set("add", func(val *goja.Object) {
			if err := t.Add(m.fromProxy(val)); err != nil {
				panic(m.runtime.ToValue(err.Error()))
			}
		})
		obj.Set("get", func(i int) *goja.Object {
			if i >= 0 && i < len(t.Value) {
				return m.toProxy(t.Value[i])
			}
			return nil
		})

		// FIX: Define 'length' and 'listType' as true getter properties.
		// This resolves the build error and fixes the logic bug.
		getterLength := m.runtime.ToValue(func() int { return len(t.Value) })
		if err := obj.DefineAccessorProperty("length", getterLength, nil, goja.FLAG_TRUE, goja.FLAG_FALSE); err != nil {
			panic(fmt.Errorf("could not define 'length' accessor: %w", err))
		}

		getterType := m.runtime.ToValue(func() nbt.TagID { return t.Type })
		if err := obj.DefineAccessorProperty("listType", getterType, nil, goja.FLAG_TRUE, goja.FLAG_FALSE); err != nil {
			panic(fmt.Errorf("could not define 'listType' accessor: %w", err))
		}

	case *nbt.ByteTag:
		obj.Set("value", t.Value)
	case *nbt.ShortTag:
		obj.Set("value", t.Value)
	case *nbt.IntTag:
		obj.Set("value", t.Value)
	case *nbt.LongTag:
		obj.Set("value", t.Value)
	case *nbt.FloatTag:
		obj.Set("value", t.Value)
	case *nbt.DoubleTag:
		obj.Set("value", t.Value)
	case *nbt.StringTag:
		obj.Set("value", t.Value)
	case *nbt.ByteArrayTag:
		obj.Set("value", t.Value)
	case *nbt.IntArrayTag:
		obj.Set("value", t.Value)
	case *nbt.LongArrayTag:
		obj.Set("value", t.Value)
	}
	return obj
}

// fromProxy extracts the Go nbt.Tag from a JavaScript proxy object.
func (m *NbtModule) fromProxy(obj *goja.Object) nbt.Tag {
	if obj == nil {
		return nil
	}
	v := obj.Get(nativeTag)
	if goja.IsUndefined(v) || goja.IsNull(v) {
		return nil
	}
	if tag, ok := v.Export().(nbt.Tag); ok {
		return tag
	}
	return nil
}

// --- Bound Functions ---

func (m *NbtModule) parseSNBT(snbt string) *goja.Object {
	tag, err := nbt.ParseSNBT(snbt)
	if err != nil {
		panic(m.runtime.ToValue(err.Error()))
	}
	return m.toProxy(tag)
}

func (m *NbtModule) toPrettySNBT(obj *goja.Object) string {
	tag := m.fromProxy(obj)
	return nbt.ToPrettySNBT(tag)
}

func (m *NbtModule) toCompactSNBT(obj *goja.Object) string {
	tag := m.fromProxy(obj)
	return nbt.ToCompactSNBT(tag)
}

// writeBinary is the internal helper for writing NBT data.
func (m *NbtModule) writeBinary(tag nbt.Tag, name string, compressed bool) goja.Value {
	var buf bytes.Buffer
	namedTag := nbt.NamedTag{Name: name, Tag: tag}
	var err error
	if compressed {
		err = nbt.WriteCompressed(&buf, namedTag)
	} else {
		err = nbt.Write(&buf, namedTag)
	}
	if err != nil {
		panic(m.runtime.ToValue(err.Error()))
	}
	return m.runtime.ToValue(m.runtime.NewArrayBuffer(buf.Bytes()))
}

// readBinary now returns a JS object: { name: string, tag: object }
func (m *NbtModule) readBinary(buffer goja.Value, compressed bool) *goja.Object {
	ab, ok := buffer.Export().(goja.ArrayBuffer)
	if !ok {
		panic(m.runtime.ToValue("expected an ArrayBuffer"))
	}

	reader := bytes.NewReader(ab.Bytes())
	var namedTag nbt.NamedTag
	var err error

	if compressed {
		namedTag, err = nbt.ReadCompressed(reader)
	} else {
		namedTag, err = nbt.Read(reader)
	}

	if err != nil {
		panic(m.runtime.ToValue(err.Error()))
	}

	result := m.runtime.NewObject()
	result.Set("name", namedTag.Name)
	result.Set("tag", m.toProxy(namedTag.Tag))

	return result
}

func (m *NbtModule) compare(objA, objB *goja.Object, partial bool) bool {
	tagA := m.fromProxy(objA)
	tagB := m.fromProxy(objB)
	return nbt.CompareTags(tagA, tagB, partial)
}
