// Helpers related to testdata assets. Only suitable for integration tests (*_test) packages
// due to circular imports
package testing

import (
	"slices"
	"strings"

	"github.com/MusicalNinjaDad/snaggle/elf"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

func AllElfs() []TestDetails {
	return filterTests(TestData, func(td TestDetails) bool { return !td.NonElf })
}

func noSubDirs() []TestDetails {
	return filterTests(TestData, func(td TestDetails) bool { return !td.InSubdir && !td.NonElf })
}

func allFiles() []TestDetails {
	return filterTests(TestData, func(_ TestDetails) bool { return true })
}

func allFilesBaseDirOnly() []TestDetails {
	return filterTests(TestData, func(td TestDetails) bool { return !td.InSubdir })
}

type TestDetails struct {
	Name           string
	Path           string
	SnagTo         string
	SnagAs         string
	InSubdir       bool
	Symlink        bool
	NonElf         bool
	Dynamic        bool
	Exe            bool
	Lib            bool
	HasInterpreter bool
	Elf            elf.Elf
}

type testListing = map[string]TestDetails

var TestData = testListing{
	P_ctypes_so: {
		Name:           "dyn_lib",
		Path:           P_ctypes_so,
		SnagTo:         "lib64",
		SnagAs:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
		Dynamic:        true,
		Exe:            false,
		Lib:            true,
		HasInterpreter: false,
		Elf: elf.Elf{
			Name:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
			Path:         P_ctypes_so,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.DYN),
			Interpreter:  "",
			Dependencies: []string{P_libc, P_libm, P_libpthread},
		},
	},
	P_empty: {
		Name:   "empty_file",
		Path:   P_empty,
		NonElf: true,
		Elf: elf.Elf{
			Name:         "empyty",
			Path:         P_empty,
			Class:        elf.EI_CLASS(elf.UNDEF),
			Type:         elf.Type(elf.ELFNONE),
			Interpreter:  "",
			Dependencies: nil,
		},
	},
	P_build_sh: {
		Path:     P_build_sh,
		InSubdir: true,
		NonElf:   true,
	},
	P_hello_dynamic: {
		Name:     "subdir",
		Path:     P_hello_dynamic,
		SnagTo:   "bin",
		SnagAs:   "hello",
		InSubdir: true,
		Elf: elf.Elf{
			Name:         "hello",
			Path:         P_hello_dynamic,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.PIE,
			Interpreter:  P_ld_linux,
			Dependencies: []string{P_libc},
		},
		Dynamic:        true,
		Exe:            true,
		Lib:            false,
		HasInterpreter: true,
	},
	P_hello_go: {
		Path:     P_hello_go,
		InSubdir: true,
		NonElf:   true,
	},
	P_hello_pie: {
		Name:   "PIE_0_deps",
		Path:   P_hello_pie,
		SnagTo: "bin",
		SnagAs: "hello_pie",
		Elf: elf.Elf{
			Name:         "hello_pie",
			Path:         P_hello_pie,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.PIE),
			Interpreter:  P_ld_linux,
			Dependencies: nil,
		},
		Dynamic:        true,
		Exe:            true,
		Lib:            false,
		HasInterpreter: true,
	},
	P_hello_static: {
		Name:   "static",
		Path:   P_hello_static,
		SnagTo: "bin",
		SnagAs: "hello_static",
		Elf: elf.Elf{
			Name:         "hello_static",
			Path:         P_hello_static,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.EXE),
			Interpreter:  "",
			Dependencies: nil,
		},
		Dynamic:        false,
		Exe:            true,
		Lib:            false,
		HasInterpreter: false,
	},
	P_id: {
		Name:   "PIE_many_deps",
		Path:   P_id,
		SnagTo: "bin",
		SnagAs: "id",
		Elf: elf.Elf{
			Name:         "id",
			Path:         P_id,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.PIE),
			Interpreter:  P_ld_linux,
			Dependencies: []string{P_libc, P_libpcre2_8, P_libselinux},
		},
		Dynamic:        true,
		Exe:            true,
		Lib:            false,
		HasInterpreter: true,
	},
	P_ldd: {
		Name:   "ldd",
		Path:   P_ldd,
		NonElf: true,
		Elf: elf.Elf{
			Name:         "ldd",
			Path:         P_ldd,
			Class:        elf.EI_CLASS(elf.UNDEF),
			Type:         elf.Type(elf.ELFNONE),
			Interpreter:  "",
			Dependencies: nil,
		},
		Dynamic:        false,
		Exe:            false,
		Lib:            false,
		HasInterpreter: false,
	},
	P_symlinked_id: {
		Name:     "symlink",
		Path:     P_symlinked_id,
		SnagTo:   "bin",
		SnagAs:   "id2",
		InSubdir: true,
		Symlink:  true,
		Elf: elf.Elf{
			Name:         "id",
			Path:         P_id,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.PIE),
			Interpreter:  P_ld_linux,
			Dependencies: []string{P_libc, P_libpcre2_8, P_libselinux},
		},
		Dynamic:        true,
		Exe:            true,
		Lib:            false,
		HasInterpreter: true,
	},
	P_which: {
		Name:   "PIE_1_dep",
		Path:   P_which,
		SnagTo: "bin",
		SnagAs: "which",
		Elf: elf.Elf{
			Name:         "which",
			Path:         P_which,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.PIE),
			Interpreter:  P_ld_linux,
			Dependencies: []string{P_libc},
		},
		Dynamic:        true,
		Exe:            true,
		Lib:            false,
		HasInterpreter: true,
	},
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
