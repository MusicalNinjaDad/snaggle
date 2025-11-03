// Helpers related to testdata assets. Only suitable for integration tests (*_test) packages
// due to circular imports
package testing

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/elf"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type binaryDetails struct {
	Description    string // test run name
	Elf            elf.Elf
	Dynamic        bool
	Exe            bool
	Lib            bool
	HasInterpreter bool
}

var commonElfs = map[string]elf.Elf{
	"hello_pie": {
		Name:         "hello_pie",
		Path:         P_hello_pie,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.PIE),
		Interpreter:  P_ld_linux,
		Dependencies: nil,
	},
	"hello_static": {
		Name:         "hello_static",
		Path:         P_hello_static,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.EXE),
		Interpreter:  "",
		Dependencies: nil,
	},
	"which": {
		Name:         "which",
		Path:         P_which,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.PIE),
		Interpreter:  P_ld_linux,
		Dependencies: []string{P_libc},
	},
	"id": {
		Name:         "id",
		Path:         P_id,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.PIE),
		Interpreter:  P_ld_linux,
		Dependencies: []string{P_libc, P_libpcre2_8, P_libselinux},
	},
	"ctypes_so": {
		Name:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
		Path:         P_ctypes_so,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "",
		Dependencies: []string{P_libc, P_libm, P_libpthread},
	},
	"hello_dynamic": {
		Name:         "hello",
		Path:         P_hello_dynamic,
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.PIE,
		Interpreter:  P_ld_linux,
		Dependencies: []string{P_libc},
	},
	"ldd": {
		Name:         "ldd",
		Path:         P_ldd,
		Class:        elf.EI_CLASS(elf.UNDEF),
		Type:         elf.Type(elf.ELFNONE),
		Interpreter:  "",
		Dependencies: nil,
	},
}

var Hello_dynamic = binaryDetails{
	Description:    "In subdirectory",
	Elf:            commonElfs["hello_dynamic"],
	Dynamic:        true,
	Exe:            true,
	Lib:            false,
	HasInterpreter: true,
}

var Ldd = binaryDetails{
	Description:    "Not an ELF",
	Elf:            commonElfs["ldd"],
	Dynamic:        false,
	Exe:            false,
	Lib:            false,
	HasInterpreter: false,
}

// Test Helper: provides a set of common binaries to use in table tests
//
//	 type binaryDetails struct {
//			Description string // test run name
//		   	Elf			elf.Elf
//		   	Dynamic     bool
//		 	Exe         bool
//			Lib         bool
//		}
func CommonBinaries(t testing.TB) map[string]binaryDetails {
	t.Helper()
	return map[string]binaryDetails{
		"PIE_0": {
			Description:    "PIE no dependencies",
			Elf:            commonElfs["hello_pie"],
			Dynamic:        true,
			Exe:            true,
			Lib:            false,
			HasInterpreter: true,
		},
		"static": {
			Description:    "Static linked executable",
			Elf:            commonElfs["hello_static"],
			Dynamic:        false,
			Exe:            true,
			Lib:            false,
			HasInterpreter: false,
		},
		"PIE_1": {
			Description:    "PIE 1 dependency",
			Elf:            commonElfs["which"],
			Dynamic:        true,
			Exe:            true,
			Lib:            false,
			HasInterpreter: true,
		},
		"PIE_Many": {
			Description:    "PIE nested dependencies",
			Elf:            commonElfs["id"],
			Dynamic:        true,
			Exe:            true,
			Lib:            false,
			HasInterpreter: true,
		},
		"dyn_lib": {
			Description:    "Dynamic library (.so)",
			Elf:            commonElfs["ctypes_so"],
			Dynamic:        true,
			Exe:            false,
			Lib:            true,
			HasInterpreter: false,
		},
	}
}

var commonBins = map[string]binaryDetails{
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

// Identify the expected outputs when snaggling a test case to tmp
func ExpectedOutput(tc binaryDetails, tmp string, inplace bool) (stdout []string, files map[string]string) {
	numFiles := 1 // snaggled Elf
	numFiles += len(tc.Elf.Dependencies)
	if tc.HasInterpreter {
		numFiles++
	}
	stdout = make([]string, 0, numFiles)
	files = make(map[string]string, numFiles)

	binBasePath := filepath.Join(tmp, "bin")
	libBasePath := filepath.Join(tmp, "lib64")

	switch inplace {
	case true:
		// Don't expect base ELF to be snagged
	case false:
		var elfPath string
		if tc.Exe {
			elfPath = filepath.Join(binBasePath, tc.Elf.Name)
		} else {
			elfPath = filepath.Join(libBasePath, tc.Elf.Name)
		}

		stdout = append(stdout, LinkMessage(tc.Elf.Path, elfPath))
		files[tc.Elf.Path] = elfPath
	}

	if tc.HasInterpreter {
		interpPath := filepath.Join(tmp, tc.Elf.Interpreter)
		stdout = append(stdout, LinkMessage(tc.Elf.Interpreter, interpPath))
		files[tc.Elf.Interpreter] = interpPath
	}

	for _, lib := range tc.Elf.Dependencies {
		libPath := filepath.Join(libBasePath, filepath.Base(lib))
		stdout = append(stdout, LinkMessage(lib, libPath))
		files[lib] = libPath
	}
	return
}

func AssertStdout(t *testing.T, expected []string, actual []string) {
	t.Helper()
	a := assert.New(t)

	stripped := make([]string, 0, len(actual))

	for n, line := range actual {
		a.Conditionf(func() (success bool) {
			return strings.HasPrefix(line, "copy ") || strings.HasPrefix(line, "link ")
		}, "Line %v does not start with `copy` or `line`: %s", n+1, line)
		stripped = append(stripped, line[5:])
	}

	a.ElementsMatch(expected, stripped)
}
