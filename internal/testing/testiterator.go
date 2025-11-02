package testing

import (
	"iter"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/MusicalNinjaDad/snaggle/elf"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
	Inplace        bool
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
		for _, inplace := range []bool{false, true} {
			for desc, bin := range tests {
				if inplace {
					desc += "_inplace"
				}
				t.Run(desc, func(t *testing.T) {
					tc := TestCase{ExpectedFiles: make(map[string]string, len(bin.elf.Dependencies)+2)}
					tc.Inplace = inplace

					tc.Src = bin.path
					tc.Dest = WorkspaceTempDir(t)

					if !inplace {
						snaggedBin := filepath.Join(tc.Dest, bin.snagto, bin.snagas)
						tc.ExpectedStdout = append(tc.ExpectedStdout,
							tc.Src+" -> "+snaggedBin,
						)
						tc.ExpectedFiles[tc.Src] = snaggedBin
					}

					if bin.hasInterpreter {
						snaggedInterp := filepath.Join(tc.Dest, bin.elf.Interpreter)
						tc.ExpectedStdout = append(tc.ExpectedStdout,
							bin.elf.Interpreter+" -> "+snaggedInterp,
						)
						tc.ExpectedFiles[bin.elf.Interpreter] = snaggedInterp
					}

					for _, lib := range bin.elf.Dependencies {
						snaggedLib := filepath.Join(tc.Dest, "lib64", filepath.Base(lib))
						tc.ExpectedStdout = append(tc.ExpectedStdout,
							lib+" -> "+snaggedLib,
						)
						tc.ExpectedFiles[lib] = snaggedLib
					}

					t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
					testbody(t, tc)
				})
			}
		}
	}
}

type testDetails struct {
	elf            elf.Elf
	path           string
	snagto         string
	snagas         string
	hasInterpreter bool
}

var tests = map[string]testDetails{
	"PIE_0_deps": {
		elf:            commonElfs["hello_pie"],
		path:           P_hello_pie,
		snagto:         "bin",
		snagas:         "hello_pie",
		hasInterpreter: true,
	},
	"static": {
		elf:            commonElfs["hello_static"],
		path:           P_hello_static,
		snagto:         "bin",
		snagas:         "hello_static",
		hasInterpreter: false,
	},
	"PIE_1_dep": {
		elf:            commonElfs["which"],
		path:           P_which,
		snagto:         "bin",
		snagas:         "which",
		hasInterpreter: true,
	},
	"PIE_many_deps": {
		elf:            commonElfs["id"],
		path:           P_id,
		snagto:         "bin",
		snagas:         "id",
		hasInterpreter: true,
	},
	"dyn_lib": {
		elf:            commonElfs["ctypes_so"],
		path:           P_ctypes_so,
		snagto:         "lib64",
		snagas:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
		hasInterpreter: false,
	},
}
