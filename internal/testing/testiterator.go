package testing

import (
	"iter"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type testListing = map[string]TestDetails

var testdata = map[string]TestDetails{
	P_ctypes_so: {
		Name:   "dyn_lib",
		Path:   P_ctypes_so,
		Bin:    GoodElfs["ctypes_so"],
		SnagTo: "lib64",
		SnagAs: "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
	},
	P_build_sh: {
		Path:     P_build_sh,
		InSubdir: true,
		NonElf:   true,
	},
	P_hello_dynamic: {
		Name:     "subdir",
		Path:     P_hello_dynamic,
		Bin:      GoodElfs["hello_dynamic"],
		SnagTo:   "bin",
		SnagAs:   "hello",
		InSubdir: true,
	},
	P_hello_go: {
		Path:     P_hello_go,
		InSubdir: true,
		NonElf:   true,
	},
	P_hello_pie: {
		Name:   "PIE_0_deps",
		Path:   P_hello_pie,
		Bin:    GoodElfs["hello_pie"],
		SnagTo: "bin",
		SnagAs: "hello_pie",
	},
	P_hello_static: {
		Name:   "static",
		Path:   P_hello_static,
		Bin:    GoodElfs["hello_static"],
		SnagTo: "bin",
		SnagAs: "hello_static",
	},
	P_id: {
		Name:   "PIE_many_deps",
		Path:   P_id,
		Bin:    GoodElfs["id"],
		SnagTo: "bin",
		SnagAs: "id",
	},
	P_ldd: {
		Name:   "ldd",
		Path:   P_ldd,
		Bin:    Ldd,
		NonElf: true,
	},
	P_symlinked_id: {
		Name:     "symlink",
		Path:     P_symlinked_id,
		Bin:      GoodElfs["id"],
		SnagTo:   "bin",
		SnagAs:   "id2",
		InSubdir: true,
		Symlink:  true,
	},
	P_which: {
		Name:   "PIE_1_dep",
		Path:   P_which,
		Bin:    GoodElfs["which"],
		SnagTo: "bin",
		SnagAs: "which",
	},
}

func defaultTests() []TestDetails {
	return filterTests(testdata, func(td TestDetails) bool { return !td.NonElf })
}

func noSubDirs() []TestDetails {
	return filterTests(testdata, func(td TestDetails) bool { return !td.InSubdir && !td.NonElf })
}

func allFiles() []TestDetails {
	return filterTests(testdata, func(_ TestDetails) bool { return true })
}

func allFilesBaseDirOnly() []TestDetails {
	return filterTests(testdata, func(td TestDetails) bool { return !td.InSubdir })
}

// Ordered lexically in alphabetical order of full path :-x
func filterTests(tests testListing, filterFunc func(TestDetails) bool) []TestDetails {
	ts := make([]TestDetails, 0)
	for _, t := range tests {
		if filterFunc(t) {
			ts = append(ts, t)
		}
	}
	slices.SortFunc(ts, func(a TestDetails, b TestDetails) int { return strings.Compare(a.Path, b.Path) })
	return ts
}

type TestDetails struct {
	Name     string
	Path     string
	Bin      binaryDetails
	SnagTo   string
	SnagAs   string
	InSubdir bool
	Symlink  bool
	NonElf   bool
}

type TestCase struct {
	Src            string            // source path
	Dest           string            // destination path
	ExpectedStdout []string          // Split by line, de-(in)dented
	ExpectedFiles  map[string]string // map[original_path]snagged_path
	Options        []snaggle.Option
	Flags          []string
}

type testOption struct {
	name           string
	negativeSuffix string
	option         snaggle.Option
	flag           string
}

type testOptions struct {
	names   []string
	options []snaggle.Option
	flags   []string
}

func combine(opts ...testOption) []testOptions {
	n := len(opts)
	options := make([]testOptions, 0, n)

	bitmask := (1 << n)
	for mask := range bitmask {
		option := testOptions{
			names:   make([]string, 0, n),
			options: make([]snaggle.Option, 0, n),
			flags:   make([]string, 0, n),
		}
		for i, opt := range opts {
			if mask&(1<<i) > 0 { //i-th bit is set
				option.names = appendif(option.names, opt.name)
				option.options = appendif(option.options, opt.option)
				option.flags = appendif(option.flags, opt.flag)
			} else {
				option.names = appendif(option.names, opt.negativeSuffix)
			}
		}
		options = append(options, option)
	}
	return options
}

func appendif[Type any](slice []Type, elem Type) []Type {
	if reflect.ValueOf(elem).IsZero() {
		return slice
	}
	return append(slice, elem)
}

var (
	no_option   = testOption{}
	copy_option = testOption{name: "copy", option: snaggle.Copy(), flag: "--copy"}
	inplace     = testOption{name: "inplace", option: snaggle.InPlace(), flag: "--in-place"}
	recursive   = testOption{name: "recursive", option: snaggle.Recursive(), flag: "--recursive"}
	relative    = testOption{name: "relative"}
	verbose     = testOption{name: "verbose", negativeSuffix: "silent", option: snaggle.Verbose(), flag: "--verbose"}
)

// Is a specific option set?
func (o *testOptions) includes(opt testOption) bool {
	return slices.Contains(o.names, opt.name)
}

// Calls t.Run on the test body for all our test case binaries e.g.:
//
//	for t, tc := range TestLoop(t) {
//		Assert := assert.New(t)
//		Assert.NotEmpty(tc)
//	}
//
// returns the specific, shadowed, t for each test run, to ensure results are correctly allocated to the subtest
//
// Just don't try to use SkipNow, FailNow or t.Parallel as Go's rangefunc & t.Run hacks collide and mess up
// goroutine clarity. There is probably a good fix with channels but I can't be bothered right now...
func TestLoop(t *testing.T, tests ...TestDetails) iter.Seq2[*testing.T, TestCase] {

	var runDirTests bool
	if tests == nil {
		tests = defaultTests()
		runDirTests = true
	} else {
		runDirTests = false
	}

	return func(testbody func(t *testing.T, tc TestCase) bool) {
		runTest := func(name string, src string, options testOptions, bins ...TestDetails) {
			var err error

			desc := strings.Join(append([]string{name}, options.names...), "_")
			t.Run(desc, func(t *testing.T) {
				tc := TestCase{
					Src:            src,
					Dest:           WorkspaceTempDir(t),
					ExpectedStdout: make([]string, 0),
					ExpectedFiles:  make(map[string]string),
					Options:        options.options,
					Flags:          options.flags,
				}
				generateOutput(&tc, options, bins...)

				if options.includes(relative) {
					tc.Dest, err = filepath.Rel(pwd(t), tc.Dest)
					assert.NoError(t, err, "conversion to relative path failed")
				}

				t.Logf("\n\nTestcase details: %s", spew.Sdump(tc))
				t.Logf("\n\nTest options: %s", spew.Sdump(options))

				testbody(t, tc)
			})
		}

		for _, options := range combine(relative, inplace, verbose) {
			for _, bin := range tests {
				runTest(bin.Name, bin.Path, options, bin)
			}
		}

		if runDirTests {
			for _, process := range []testOption{no_option, copy_option, inplace} {
				for _, options := range combine(relative, verbose, recursive) {

					options.names = appendif(options.names, process.name)
					options.options = appendif(options.options, process.option)
					options.flags = appendif(options.flags, process.flag)

					var bins []TestDetails
					switch {
					case options.includes(copy_option) && options.includes(recursive):
						bins = allFiles()
					case options.includes(copy_option):
						bins = allFilesBaseDirOnly()
					case options.includes(recursive):
						bins = defaultTests()
					default:
						bins = noSubDirs()
					}

					runTest("directory", TestdataPath("."), options, bins...)
				}
			}
		}
	}
}

func generateOutput(tc *TestCase, options testOptions, bins ...TestDetails) {
	for _, bin := range bins {
		var copy_bin bool
		var snaggedBin string
		switch {
		case bin.NonElf && !options.includes(copy_option):
			copy_bin = false
		case options.includes(inplace):
			copy_bin = false
		case options.includes(copy_option):
			copy_bin = true
			snaggedBin = filepath.Join(tc.Dest, bin.Path)
		default:
			copy_bin = true
			snaggedBin = filepath.Join(tc.Dest, bin.SnagTo, bin.SnagAs)
		}

		if copy_bin {
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
	if options.includes(verbose) {
		// we want stdout
	} else {
		tc.ExpectedStdout = make([]string, 0)
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
