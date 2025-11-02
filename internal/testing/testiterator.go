package testing

import (
	"iter"
	"maps"
	"path/filepath"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/MusicalNinjaDad/snaggle"
	"github.com/MusicalNinjaDad/snaggle/elf"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
	Options        []snaggle.Option
	Flags          []string
}

// Calls t.Run on the test body for all our test case binaries e.g.:
//
//	for t, tc := range TestCases(t) {
//		Assert := assert.New(t)
//		Assert.NotEmpty(tc)
//	}
//
// returns the specific, shadowed, t for each test run, to ensure results are correctly allocated to the subtest
func TestCases(t *testing.T) iter.Seq2[*testing.T, TestCase] {
	return func(testbody func(t *testing.T, tc TestCase) bool) {
		for _, inplace := range []bool{false, true} {
			for desc, bin := range tests {

				var options []snaggle.Option
				var flags []string
				if inplace {
					desc += "_inplace"
					flags = append(flags, "--in-place")
					options = append(options, snaggle.InPlace())
				}

				t.Run(desc, func(t *testing.T) {
					tc := TestCase{
						Src:            bin.path,
						Dest:           WorkspaceTempDir(t),
						ExpectedStdout: make([]string, 0, len(bin.elf.Dependencies)+2),
						ExpectedFiles:  make(map[string]string, len(bin.elf.Dependencies)+2),
						Options:        options,
						Flags:          flags,
					}

					generateOutput(bin, &tc, inplace)

					t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
					testbody(t, tc)
				})
			}

			for _, recursive := range []bool{false, true} {
				desc := "Directory"
				bins := maps.Clone(tests)

				var options []snaggle.Option
				var flags []string
				if inplace {
					desc += "_inplace"
					options = append(options, snaggle.InPlace())
					flags = append(flags, "--in-place")
				}
				if recursive {
					desc += "_recursive"
					options = append(options, snaggle.Recursive())
					flags = append(flags, "--recursive")
					bins["subdir"] = subdir_contents
				}

				t.Run(desc, func(t *testing.T) {
					tc := TestCase{
						Src:            TestdataPath("."),
						Dest:           WorkspaceTempDir(t),
						ExpectedStdout: make([]string, 0),
						ExpectedFiles:  make(map[string]string),
						Options:        options,
						Flags:          flags,
					}

					for _, bin := range bins {
						generateOutput(bin, &tc, inplace)
					}

					t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
					testbody(t, tc)

				})

			}
		}
	}
}

func generateOutput(bin testDetails, tc *TestCase, inplace bool) {
	if !inplace {
		snaggedBin := filepath.Join(tc.Dest, bin.snagto, bin.snagas)
		tc.ExpectedStdout = append(tc.ExpectedStdout,
			bin.path+" -> "+snaggedBin,
		)
		tc.ExpectedFiles[bin.path] = snaggedBin
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

var subdir_contents = testDetails{
	elf:            commonElfs["hello_dynamic"],
	path:           P_hello_dynamic,
	snagto:         "bin",
	snagas:         "hello",
	hasInterpreter: true,
}
