package tests

import (
	"github.com/Advik-B/Golem/javascript"
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

// testReporter allows JS to report failures back to the Go test runner.
type testReporter struct {
	t    *testing.T
	mu   sync.Mutex
	fail bool
}

func (r *testReporter) Fail(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.t.Errorf("JS test failure: %s", msg)
	r.fail = true
}

func (r *testReporter) Log(msg string) {
	r.t.Logf("JS log: %s", msg)
}

// TestGojaNBTBindings runs all JavaScript-based tests for the NBT bindings.
func TestGojaNBTBindings(t *testing.T) {
	tests := map[string]string{
		"SNBT Parsing and Printing": `
			const snbt = '{ "key": "value", list: [1b, 2b, 3b] }';
			const tag = nbt.parse(snbt);
			if (tag.get('key').value !== 'value') test.fail("Parsed string value is incorrect");
			if (tag.get('list').length !== 3) test.fail("Parsed list length is incorrect");

			const pretty = tag.toPretty();
			if (!pretty.includes('"key": "value"')) test.fail("Pretty print is missing key-value");
			test.log("Pretty SNBT:\n" + pretty);
		`,
		"Binary Round Trip": `
			const compound = nbt.newCompound();
			compound.set("long", nbt.newLong(1234567890));
			compound.set("string", nbt.newString("hello world"));

			const binary = compound.write("rootName");
			const result = nbt.read(binary);
			
			if (result.name !== "rootName") test.fail("Root name was lost in translation. Got: " + result.name);
			if (!nbt.compare(compound, result.tag, false)) test.fail("Tag is not the same after round trip");
		`,
		"Compressed Round Trip": `
			const tag = nbt.newIntArray([10, 20, 30]);
			const compressedBinary = tag.writeCompressed("compressedRoot");
			const result = nbt.readCompressed(compressedBinary);

			if (result.name !== "compressedRoot") test.fail("Root name was lost in compressed translation.");
			if (!nbt.compare(tag, result.tag, false)) test.fail("Tag is not the same after compressed round trip");
		`,
		"Tag Creation and Getters": `
			const comp = nbt.newCompound();
			comp.set("b", nbt.newByte(127));
			if (comp.get("b").value !== 127) test.fail("Byte value mismatch.");
			if (comp.get("b").typeName() !== "TAG_Byte") test.fail("Byte type name mismatch.");
		`,
		"List Operations": `
			const list = nbt.newList();
			list.add(nbt.newInt(1));
			list.add(nbt.newInt(2));
			if (list.length !== 2) test.fail("List length should be 2");
			if (list.get(1).value !== 2) test.fail("List element value is incorrect");
			if (list.listType !== 3) test.fail("List type ID should be TAG_Int (3)");
		`,
		"Error on Invalid SNBT": `
			try {
				nbt.parse('{key:"value",}'); // trailing comma
				test.fail("Parsing invalid SNBT should have thrown an error.");
			} catch (e) {
				test.log("Successfully caught expected error: " + e);
			}
		`,
		"Error on Mismatched List Add": `
			try {
				const list = nbt.newList();
				list.add(nbt.newInt(1));
				list.add(nbt.newString("I should not be here")); // Mismatched type
				test.fail("Adding a mismatched type to a list should have thrown an error.");
			} catch (e) {
				test.log("Successfully caught expected error: " + e);
			}
		`,
	}

	for name, script := range tests {
		t.Run(name, func(t *testing.T) {
			vm := goja.New()
			reporter := &testReporter{t: t}

			// Inject the test reporter
			vm.Set("test", map[string]interface{}{
				"fail": reporter.Fail,
				"log":  reporter.Log,
			})

			// Inject the NBT module
			nbtModule := javascript.NewNbtModule(vm)
			vm.Set("nbt", nbtModule)

			// Run the test script
			_, err := vm.RunString(script)

			// Check for Go-level errors (panics from the bindings)
			require.NoError(t, err, "Goja script execution failed")
			// Check for JS-level failures reported via test.fail()
			require.False(t, reporter.fail, "One or more JavaScript assertions failed")
		})
	}
}
