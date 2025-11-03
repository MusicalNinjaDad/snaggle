package testing

import (
	"iter"
	"path/filepath"
	"slices"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/MusicalNinjaDad/snaggle"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

var defaultTests = []TestDetails{
	{
		Name:   "PIE_0_deps",
		Path:   P_hello_pie,
		Bin:    GoodElfs["hello_pie"],
		SnagTo: "bin",
		SnagAs: "hello_pie",
	},
	{
		Name:   "static",
		Path:   P_hello_static,
		Bin:    GoodElfs["hello_static"],
		SnagTo: "bin",
		SnagAs: "hello_static",
	},
	{
		Name:   "PIE_1_dep",
		Path:   P_which,
		Bin:    GoodElfs["which"],
		SnagTo: "bin",
		SnagAs: "which",
	},
	{
		Name:   "PIE_many_deps",
		Path:   P_id,
		Bin:    GoodElfs["id"],
		SnagTo: "bin",
		SnagAs: "id",
	},
	{
		Name:   "dyn_lib",
		Path:   P_ctypes_so,
		Bin:    GoodElfs["ctypes_so"],
		SnagTo: "lib64",
		SnagAs: "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
	},
}

type TestDetails struct {
	Name   string
	Path   string
	Bin    binaryDetails
	SnagTo string
	SnagAs string
}

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
	ExpectedBinary binaryDetails
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
//
// Just don't try to use SkipNow, FailNow or t.Parallel as Go's rangefunc & t.Run hacks collide and mess up
// goroutine clarity. There is probably a good fix with channels but I can't be bothered right now...
func TestCases(t *testing.T) iter.Seq2[*testing.T, TestCase] {
	return func(testbody func(t *testing.T, tc TestCase) bool) {
		for _, inplace := range []bool{false, true} {
			for _, bin := range defaultTests {

				desc := bin.Name

				var options []snaggle.Option
				var flags []string
				if inplace {
					desc += "_inplace"
					flags = append(flags, "--in-place")
					options = append(options, snaggle.InPlace())
				}

				t.Run(desc, func(t *testing.T) {
					tc := TestCase{
						Src:            bin.Path,
						Dest:           WorkspaceTempDir(t),
						ExpectedStdout: make([]string, 0, len(bin.Bin.Elf.Dependencies)+2),
						ExpectedFiles:  make(map[string]string, len(bin.Bin.Elf.Dependencies)+2),
						ExpectedBinary: bin.Bin,
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
				bins := slices.Clone(defaultTests)

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
					bins = append(bins, subdir_contents)
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

func generateOutput(bin TestDetails, tc *TestCase, inplace bool) {
	if !inplace {
		snaggedBin := filepath.Join(tc.Dest, bin.SnagTo, bin.SnagAs)
		tc.ExpectedStdout = append(tc.ExpectedStdout,
			bin.Path+" -> "+snaggedBin,
		)
		tc.ExpectedFiles[bin.Path] = snaggedBin
	}

	if bin.Bin.HasInterpreter {
		snaggedInterp := filepath.Join(tc.Dest, bin.Bin.Elf.Interpreter)
		tc.ExpectedStdout = append(tc.ExpectedStdout,
			bin.Bin.Elf.Interpreter+" -> "+snaggedInterp,
		)
		tc.ExpectedFiles[bin.Bin.Elf.Interpreter] = snaggedInterp
	}

	for _, lib := range bin.Bin.Elf.Dependencies {
		snaggedLib := filepath.Join(tc.Dest, "lib64", filepath.Base(lib))
		tc.ExpectedStdout = append(tc.ExpectedStdout,
			lib+" -> "+snaggedLib,
		)
		tc.ExpectedFiles[lib] = snaggedLib
	}
}

var subdir_contents = TestDetails{
	Path:   P_hello_dynamic,
	Bin:    GoodElfs["hello_dynamic"],
	SnagTo: "bin",
	SnagAs: "hello",
}
