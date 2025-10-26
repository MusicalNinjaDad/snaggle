// Helpers related to testdata assets. Only suitable for integration tests (*_test) packages
// due to circular imports
package testing

//nolint:staticcheck
import (
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"

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

// Identify the expected Stdout when snaggling a test case to tmp
func ExpectedOutput(tc binaryDetails, tmp string) (stdout []string) {
	files := 1 // snaggled Elf
	files += len(tc.Elf.Dependencies)
	if tc.HasInterpreter {
		files++
	}
	stdout = make([]string, 0, files)
	binPath := filepath.Join(tmp, "bin")
	stdout = append(stdout, LinkMessage(tc.Elf.Path, filepath.Join(binPath, tc.Elf.Name)))
	if tc.HasInterpreter {
		stdout = append(stdout, LinkMessage(tc.Elf.Interpreter, filepath.Join(tmp, tc.Elf.Interpreter)))
	}
	libPath := filepath.Join(tmp, "lib64")
	for _, lib := range tc.Elf.Dependencies {
		stdout = append(stdout, LinkMessage(lib, filepath.Join(libPath, filepath.Base(lib))))
	}
	return
}
