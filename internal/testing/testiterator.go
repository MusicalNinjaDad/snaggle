package testing

import (
	"iter"
	"testing"

	"github.com/davecgh/go-spew/spew"

	//nolint:staticcheck
	"github.com/MusicalNinjaDad/snaggle/elf"
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
}

// Calls t.Run on the test body for all our test case binaries e.g.:
//
//	for t, tc := range TestCases(t) {
//		Assert := assert.New(t)
//		Assert.NotEmpty(tc)
//	}
//
// returns the sepecific, shadowed, t for each test run, to ensure results are correctly allocated to the subtest
func TestCases(t *testing.T) iter.Seq2[*testing.T, TestCase] {
	return func(testbody func(t *testing.T, tc TestCase) bool) {
		for desc, bin := range tests {
			t.Run(desc, func(t *testing.T) {
				tc := TestCase{}

				tc.Src = bin.Path
				tc.Dest = WorkspaceTempDir(t)

				t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
				testbody(t, tc)
			})
		}
	}
}

type testDetails struct {
	Elf    elf.Elf
	Path   string
	snagto string
}

var tests = map[string]testDetails{
	"PIE_0_deps": {
		Elf:    commonElfs["hello_pie"],
		Path:   P_hello_pie,
		snagto: "bin",
	},
	"static": {
		Elf:    commonElfs["hello_static"],
		Path:   P_hello_static,
		snagto: "bin",
	},
	"PIE_1_dep": {
		Elf:    commonElfs["which"],
		Path:   P_which,
		snagto: "bin",
	},
	"PIE_many_deps": {
		Elf:    commonElfs["id"],
		Path:   P_id,
		snagto: "bin",
	},
	"dyn_lib": {
		Elf:    commonElfs["ctypes_so"],
		Path:   P_ctypes_so,
		snagto: "lib64",
	},
}
