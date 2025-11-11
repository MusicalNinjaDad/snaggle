// Helpers related to testdata assets. Only suitable for integration tests (*_test) packages
// due to circular imports
package testing

import (
	"github.com/MusicalNinjaDad/snaggle/elf"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type binaryDetails struct {
	Description    string
	Elf            elf.Elf
	Dynamic        bool
	Exe            bool
	Lib            bool
	HasInterpreter bool
}

var GoodElfs = map[string]binaryDetails{
	"hello_pie": {
		Description: "PIE no dependencies",
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
	"hello_static": {
		Description: "Static linked executable",
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
	"which": {
		Description: "PIE 1 dependency",
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
	"id": {
		Description: "PIE nested dependencies",
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
	"ctypes_so": {
		Description: "Dynamic library (.so)",
		Elf: elf.Elf{
			Name:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
			Path:         P_ctypes_so,
			Class:        elf.EI_CLASS(elf.ELF64),
			Type:         elf.Type(elf.DYN),
			Interpreter:  "",
			Dependencies: []string{P_libc, P_libm, P_libpthread},
		},
		Dynamic:        true,
		Exe:            false,
		Lib:            true,
		HasInterpreter: false,
	},
	"hello_dynamic": {
		Description: "In subdirectory",
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
}

var Ldd = binaryDetails{
	Description: "Not an ELF",
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
