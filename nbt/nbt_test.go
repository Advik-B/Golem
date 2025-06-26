package nbt

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestTag is a helper function to build a complex CompoundTag for use in multiple tests.
// It contains at least one of every tag type.
func createTestTag() *CompoundTag {
	doubleList := &ListTag{Type: TagDouble}
	_ = doubleList.Add(&DoubleTag{Value: 1.1})
	_ = doubleList.Add(&DoubleTag{Value: 2.2})
	_ = doubleList.Add(&DoubleTag{Value: 3.3})

	nestedCompound := NewCompoundTag()
	nestedCompound.Put("name", &StringTag{Value: "Nested Tag"})
	nestedCompound.Put("created", &LongTag{Value: 1234567890})

	master := NewCompoundTag()
	master.Put("byte_tag", &ByteTag{Value: 127})
	master.Put("negative_byte_tag", &ByteTag{Value: -128})
	master.Put("short_tag", &ShortTag{Value: 32767})
	master.Put("int_tag", &IntTag{Value: 2147483647})
	master.Put("long_tag", &LongTag{Value: 9223372036854775807})
	master.Put("float_tag", &FloatTag{Value: 3.14159})
	master.Put("double_tag", &DoubleTag{Value: 2.71828})
	master.Put("string_tag", &StringTag{Value: "Hello, World! This string has \"quotes\" and 'apostrophes'."})
	master.Put("byte_array_tag", &ByteArrayTag{Value: []int8{0, 1, 2, 3, 4, 5}})
	master.Put("int_array_tag", &IntArrayTag{Value: []int32{-10, 0, 10, 20}})
	master.Put("long_array_tag", &LongArrayTag{Value: []int64{-1000, 0, 1000, 2000}})
	master.Put("list_of_doubles", doubleList)
	master.Put("nested_compound", nestedCompound)
	master.Put("empty_list", &ListTag{})
	master.Put("empty_compound", NewCompoundTag())

	return master
}

// TestBinaryIO covers standard and GZIP-compressed reading and writing.
// This version adds detailed logging on failure.
func TestBinaryIO(t *testing.T) {
	originalTag := createTestTag()
	namedTag := NamedTag{Name: "Master Test", Tag: originalTag}

	t.Run("Standard", func(t *testing.T) {
		var buf bytes.Buffer
		err := Write(&buf, namedTag)
		require.NoError(t, err, "Standard binary write should not fail")

		readTag, err := Read(&buf)
		require.NoError(t, err, "Standard binary read should not fail")

		// Use assert.True with verbose logging on failure
		if !assert.True(t, CompareTags(originalTag, readTag.Tag, false), "Standard binary read tag should match original") {
			t.Logf("--- VERBOSE FAILURE INFO ---")
			t.Logf("Original Tag SNBT:\n%s\n", ToPrettySNBT(originalTag))
			t.Logf("Tag Read from Binary SNBT:\n%s\n", ToPrettySNBT(readTag.Tag))
			t.Logf("--- END VERBOSE INFO ---")
		}
	})

	t.Run("Compressed", func(t *testing.T) {
		var buf bytes.Buffer
		err := WriteCompressed(&buf, namedTag)
		require.NoError(t, err, "Compressed binary write should not fail")

		readTag, err := ReadCompressed(&buf)
		require.NoError(t, err, "Compressed binary read should not fail")

		if !assert.True(t, CompareTags(originalTag, readTag.Tag, false), "Compressed binary read tag should match original") {
			t.Logf("--- VERBOSE FAILURE INFO ---")
			t.Logf("Original Tag SNBT:\n%s\n", ToPrettySNBT(originalTag))
			t.Logf("Tag Read from Compressed Binary SNBT:\n%s\n", ToPrettySNBT(readTag.Tag))
			t.Logf("--- END VERBOSE INFO ---")
		}
	})
}

// TestSNBTParsing ensures the string-to-tag parser works for valid and invalid inputs.
func TestSNBTParsing(t *testing.T) {
	t.Run("ValidSNBT", func(t *testing.T) {
		snbtString := `
		{
			"unquoted_key": "some value",
			'quoted_key': 123b,
			true_val: true,
			false_val: false,
			num_list: [1.0d, 2.0d, 3.0d],
			byte_array: [B; 1, 2, 3],
			int_array: [I; 10, 20, 30],
			long_array: [L; 100, 200, 300]
		}`

		parsed, err := ParseSNBT(snbtString)
		require.NoError(t, err, "Should successfully parse valid SNBT")
		require.NotNil(t, parsed)

		// FIX: Update assertions to expect int8 and []int8 types.
		s, ok := parsed.GetString("unquoted_key")
		assert.True(t, ok)
		assert.Equal(t, "some value", s)

		bTag, ok := parsed.Get("quoted_key")
		require.True(t, ok)
		assert.Equal(t, int8(123), bTag.(*ByteTag).Value)

		trueVal, ok := parsed.Get("true_val")
		require.True(t, ok)
		assert.Equal(t, int8(1), trueVal.(*ByteTag).Value)

		baTag, ok := parsed.Get("byte_array")
		require.True(t, ok)
		assert.Equal(t, []int8{1, 2, 3}, baTag.(*ByteArrayTag).Value)

		listTag, ok := parsed.GetList("num_list")
		require.True(t, ok)
		assert.Equal(t, TagDouble, listTag.Type, "List type should be correctly identified as double")
		assert.Len(t, listTag.Value, 3, "List should have 3 elements")
	})

	t.Run("InvalidSNBT", func(t *testing.T) {
		// Trailing comma
		_, err := ParseSNBT(`{key:"value",}`)
		assert.Error(t, err, "Should fail parsing SNBT with a trailing comma")

		// Mismatched brackets
		_, err = ParseSNBT(`{key:"value"`)
		assert.Error(t, err, "Should fail parsing SNBT with mismatched brackets")

		// Mixed list types
		_, err = ParseSNBT(`[1, "hello"]`)
		assert.Error(t, err, "Should fail parsing a list with mixed types")
	})
}

// TestSNBTPrinting ensures the tag-to-string printer works and is consistent with the parser.
func TestSNBTPrintingAndRoundTrip(t *testing.T) {
	originalTag := createTestTag()

	snbtOut := ToPrettySNBT(originalTag)
	require.NotEmpty(t, snbtOut, "Generated SNBT should not be empty")
	t.Logf("Generated SNBT for Round-Trip Test:\n%s", snbtOut)

	parsedBack, err := ParseSNBT(snbtOut)
	require.NoError(t, err, "Should be able to parse back the generated SNBT")

	if !assert.True(t, CompareTags(originalTag, parsedBack, false), "SNBT round-trip should produce an identical tag") {
		t.Logf("--- VERBOSE FAILURE INFO ---")
		t.Logf("Original Tag SNBT:\n%s\n", ToPrettySNBT(originalTag))
		t.Logf("Parsed-Back Tag SNBT:\n%s\n", ToPrettySNBT(parsedBack))
		t.Logf("--- END VERBOSE INFO ---")
	}
}

// TestCompareTags specifically tests the tag comparison utility.
func TestCompareTags(t *testing.T) {
	masterTag := createTestTag()

	t.Run("Identical", func(t *testing.T) {
		tagA := masterTag.Copy()
		tagB := masterTag.Copy()
		assert.True(t, CompareTags(tagA, tagB, false), "Two identical tags should be equal")
	})

	t.Run("Different", func(t *testing.T) {
		tagA := masterTag.Copy().(*CompoundTag)
		tagB := masterTag.Copy().(*CompoundTag)
		tagB.Put("int_tag", &IntTag{Value: 0})
		assert.False(t, CompareTags(tagA, tagB, false), "Two different tags should not be equal")
	})

	t.Run("ValidSubset", func(t *testing.T) {
		subsetTag := NewCompoundTag()
		subsetTag.Put("int_tag", &IntTag{Value: 2147483647})
		assert.True(t, CompareTags(subsetTag, masterTag, true), "A valid subset should pass partial comparison")
	})

	t.Run("InvalidSubset", func(t *testing.T) {
		subsetTag := NewCompoundTag()
		subsetTag.Put("non_existent_key", &ByteTag{Value: 1})
		assert.False(t, CompareTags(subsetTag, masterTag, true), "An invalid subset should fail partial comparison")
	})
}

// TestStructureUtils verifies the special packing/unpacking logic for structure templates.
func TestStructureUtils(t *testing.T) {
	structureSnbt := `
	{
		palette:[
			{Name:"minecraft:air"},
			{Name:"minecraft:stone",Properties:{polished:"true"}}
		],
		blocks:[
			{pos:[I; 0, 0, 0], state:0},
			{pos:[I; 1, 2, 3], state:1, nbt:{some_data:1b}}
		]
	}`

	structureTag, err := ParseSNBT(structureSnbt)
	require.NoError(t, err, "Failed to parse sample structure SNBT")

	packedSnbt, err := StructureToSnbt(structureTag)
	require.NoError(t, err, "StructureToSnbt failed")

	t.Logf("Packed Structure SNBT:\n%s", packedSnbt)

	unpackedTag, err := SnbtToStructure(packedSnbt)
	require.NoError(t, err, "SnbtToStructure failed")

	// Verify the unpacked structure is correct
	blocks, ok := unpackedTag.GetList("blocks")
	require.True(t, ok)
	require.Len(t, blocks.Value, 2)

	block1, ok := blocks.Value[1].(*CompoundTag)
	require.True(t, ok)

	state, ok := block1.GetInt("state")
	require.True(t, ok)
	assert.Equal(t, int32(1), state)

	palette, ok := unpackedTag.GetList("palette")
	require.True(t, ok)
	require.Len(t, palette.Value, 2)

	stoneState, ok := palette.Value[1].(*CompoundTag)
	require.True(t, ok)

	name, ok := stoneState.GetString("Name")
	require.True(t, ok)
	assert.Equal(t, "minecraft:stone", name)
}
