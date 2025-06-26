package js

import (
	_ "embed" // For embedding test scripts
	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
	"sync"
	"testing"
)

//go:embed testdata/snbt_parsing.js
var snbtParsing string

//go:embed testdata/binary_round_trip.js
var binaryRoundTrip string

//go:embed testdata/compressed_round_trip.js
var compressedRoundTrip string

//go:embed testdata/tag_creation_and_getters.js
var tagCreationAndGetters string

//go:embed testdata/list_operations.js
var listOperations string

//go:embed testdata/error_on_invalid_snbt.js
var errorOnInvalidSNBT string

//go:embed testdata/error_on_mismatched_list_add.js
var errorOnMismatchedListAdd string

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

// TestGojaNBTBindings runs all JavaScript-based testdata for the NBT bindings.
func TestGojaNBTBindings(t *testing.T) {
	tests := map[string]string{
		"SNBT Parsing and Printing":    snbtParsing,
		"Binary Round Trip":            binaryRoundTrip,
		"Compressed Round Trip":        compressedRoundTrip,
		"Tag Creation and Getters":     tagCreationAndGetters,
		"List Operations":              listOperations,
		"Error on Invalid SNBT":        errorOnInvalidSNBT,
		"Error on Mismatched List Add": errorOnMismatchedListAdd,
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
			nbtModule := NewNbtModule(vm)
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
