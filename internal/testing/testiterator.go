package testing

import (
	"iter"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/MusicalNinjaDad/snaggle"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

// Ordered lexically in alphabetical order of full path :-x
var defaultTests = []TestDetails{
	{
		Name:   "dyn_lib",
		Path:   P_ctypes_so,
		Bin:    GoodElfs["ctypes_so"],
		SnagTo: "lib64",
		SnagAs: "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
	},
	{
		Name:     "subdir",
		Path:     P_hello_dynamic,
		Bin:      GoodElfs["hello_dynamic"],
		SnagTo:   "bin",
		SnagAs:   "hello",
		InSubdir: true,
	},
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
		Name:   "PIE_many_deps",
		Path:   P_id,
		Bin:    GoodElfs["id"],
		SnagTo: "bin",
		SnagAs: "id",
	},
	{
		Name:     "symlink",
		Path:     P_symlinked_id,
		Bin:      GoodElfs["id"],
		SnagTo:   "bin",
		SnagAs:   "id2",
		InSubdir: true,
		Symlink:  true,
	},
	{
		Name:   "PIE_1_dep",
		Path:   P_which,
		Bin:    GoodElfs["which"],
		SnagTo: "bin",
		SnagAs: "which",
	},
}

type TestDetails struct {
	Name     string
	Path     string
	Bin      binaryDetails
	SnagTo   string
	SnagAs   string
	InSubdir bool
	Symlink  bool
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
func TestCases(t *testing.T, tests ...TestDetails) iter.Seq2[*testing.T, TestCase] {
	var specficTestsRequested bool
	if tests == nil {
		tests = slices.Clone(defaultTests)
		specficTestsRequested = false
	} else {
		specficTestsRequested = true
	}

	return func(testbody func(t *testing.T, tc TestCase) bool) {
		for _, relative := range []bool{false, true} {
			for _, inplace := range []bool{false, true} {
				for _, verbose := range []bool{false, true} {
					for _, bin := range tests {

						desc := bin.Name

						var options []snaggle.Option
						var flags []string
						if relative {
							desc += "_relative"
						}

						if inplace {
							desc += "_inplace"
							flags = append(flags, "--in-place")
							options = append(options, snaggle.InPlace())
						}

						if verbose {
							desc += "_verbose"
							flags = append(flags, "--verbose")
							options = append(options, snaggle.Verbose())
						} else {
							desc += "_silent"
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

							if relative {
								wd := pwd(t)
								rel, err := filepath.Rel(wd, tc.Dest)
								if err != nil {
									t.Errorf("Unable to define %s relative to %s", rel, wd)
								}
								tc.Dest = rel
							}

							if !verbose {
								tc.ExpectedStdout = make([]string, 0)
							}

							t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
							testbody(t, tc)
						})
					}

					if specficTestsRequested {
						continue
					}
					// else default test run includes snaggling a directory
					for _, recursive := range []bool{false, true} {
						desc := "Directory"

						var bins []TestDetails
						var options []snaggle.Option
						var flags []string

						if relative {
							desc += "_relative"
						}

						if inplace {
							desc += "_inplace"
							options = append(options, snaggle.InPlace())
							flags = append(flags, "--in-place")
						}

						if recursive {
							desc += "_recursive"
							options = append(options, snaggle.Recursive())
							flags = append(flags, "--recursive")
							bins = slices.Clone(defaultTests)
						} else {
							for _, bin := range defaultTests {
								if !bin.InSubdir {
									bins = append(bins, bin)
								}
							}
						}

						if verbose {
							desc += "_verbose"
							flags = append(flags, "--verbose")
							options = append(options, snaggle.Verbose())
						} else {
							desc += "_silent"
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

							if relative {
								wd := pwd(t)
								rel, err := filepath.Rel(wd, tc.Dest)
								if err != nil {
									t.Errorf("Unable to define %s relative to %s", rel, wd)
								}
								tc.Dest = rel
							}

							if !verbose {
								tc.ExpectedStdout = make([]string, 0)
							}

							t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
							testbody(t, tc)

						})

					}

					// and copy directory
					for _, recursive := range []bool{false, true} {
						desc := "Directory_copy"

						var bins []TestDetails
						var options []snaggle.Option
						var flags []string

						options = append(options, snaggle.Copy())

						if relative {
							desc += "_relative"
						}

						if inplace {
							desc += "_inplace"
							options = append(options, snaggle.InPlace())
							flags = append(flags, "--in-place")
						}

						if recursive {
							desc += "_recursive"
							options = append(options, snaggle.Recursive())
							flags = append(flags, "--recursive")
							bins = slices.Clone(defaultTests)
						} else {
							for _, bin := range defaultTests {
								if !bin.InSubdir {
									bins = append(bins, bin)
								}
							}
						}

						if verbose {
							desc += "_verbose"
							flags = append(flags, "--verbose")
							options = append(options, snaggle.Verbose())
						} else {
							desc += "_silent"
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

							if relative {
								wd := pwd(t)
								rel, err := filepath.Rel(wd, tc.Dest)
								if err != nil {
									t.Errorf("Unable to define %s relative to %s", rel, wd)
								}
								tc.Dest = rel
							}

							if !verbose {
								tc.ExpectedStdout = make([]string, 0)
							}

							t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
							testbody(t, tc)

						})

					}

				}
			}
		}
	}
}

func generateOutput(bin TestDetails, tc *TestCase, inplace bool) {
	if !inplace {
		snaggedBin := filepath.Join(tc.Dest, bin.SnagTo, bin.SnagAs)
		if bin.Symlink {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				bin.Path+" ("+bin.Bin.Elf.Path+") -> "+snaggedBin,
			)
		} else {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				bin.Path+" -> "+snaggedBin,
			)
		}
		tc.ExpectedFiles[bin.Path] = snaggedBin
	}

	if bin.Bin.HasInterpreter {
		snaggedInterp := filepath.Join(tc.Dest, bin.Bin.Elf.Interpreter)
		if P_ld_linux != P_ld_linux_resolved {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				bin.Bin.Elf.Interpreter+" ("+P_ld_linux_resolved+") -> "+snaggedInterp,
			)
		} else {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				bin.Bin.Elf.Interpreter+" -> "+snaggedInterp,
			)
		}
		tc.ExpectedFiles[bin.Bin.Elf.Interpreter] = snaggedInterp
	}

	for _, lib := range bin.Bin.Elf.Dependencies {
		snaggedLib := filepath.Join(tc.Dest, "lib64", filepath.Base(lib))
		resolved, _ := filepath.EvalSymlinks(lib)
		if lib != resolved {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				lib+" ("+resolved+") -> "+snaggedLib,
			)
		} else {
			tc.ExpectedStdout = append(tc.ExpectedStdout,
				lib+" -> "+snaggedLib,
			)
		}
		tc.ExpectedFiles[lib] = snaggedLib
	}
}

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Error("Failed to get pwd. Error:", err)
	}
	return pwd
}
