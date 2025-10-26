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
	Description string // test run name
	Elf         elf.Elf
	Dynamic     bool
	Exe         bool
	Lib         bool
}

// Test Helper: provides a set of common binaries to use in table tests
//
//	 type binaryDetails struct {
//			Description string // test run name
//		   	ExpectedElf elf.Elf
//		   	Dynamic     bool
//		 	Exe         bool
//			Lib         bool
//		}
func CommonBinaries(t testing.TB) map[string]binaryDetails {
	t.Helper()
	return map[string]binaryDetails{
		"PIE_0": {
			Description: "PIE no dependencies",
			Elf: elf.Elf{
				Name:         "hello_pie",
				Path:         P_hello_pie,
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: nil,
			},
			Dynamic: true,
			Exe:     true,
			Lib:     false,
		},
		"static": {
			Description: "Static linked executable",
			Elf: elf.Elf{
				Name:         "hello_static",
				Path:         P_hello_static,
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.EXE),
				Interpreter:  "",
				Dependencies: nil,
			},
			Dynamic: false,
			Exe:     true,
			Lib:     false,
		},
		"PIE_1": {
			Description: "PIE 1 dependency",
			Elf: elf.Elf{
				Name:         "which",
				Path:         P_which,
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: []string{P_libc},
			},
			Dynamic: true,
			Exe:     true,
			Lib:     false,
		},
		"PIE_Many": {
			Description: "PIE nested dependencies",
			Elf: elf.Elf{
				Name:         "id",
				Path:         P_id,
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: []string{P_libc, P_libpcre2_8, P_libselinux},
			},
			Dynamic: true,
			Exe:     true,
			Lib:     false,
		},
		"dyn_lib": {
			Description: "Dynamic library (.so)",
			Elf: elf.Elf{
				Name:         "_ctypes_test.cpython-314-x86_64-linux-gnu.so",
				Path:         P_ctypes_so,
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.DYN),
				Interpreter:  "",
				Dependencies: []string{P_libc, P_libm, P_libpthread},
			},
			Dynamic: true,
			Exe:     false,
			Lib:     true,
		},
	}
}

// Identify the expected Stdout when snaggling a test case to tmp
func ExpectedOutput(tc binaryDetails, tmp string) (stdout []string) {
	files := 1 // snaggled Elf
	files += len(tc.Elf.Dependencies)
	if tc.Elf.Type == elf.PIE {
		files++ // interpreter
	}
	stdout = make([]string, 0, files)
	binPath := filepath.Join(tmp, "bin")
	stdout = append(stdout, LinkMessage(tc.Elf.Path, filepath.Join(binPath, tc.Elf.Name)))
	if tc.Elf.Type == elf.PIE {
		stdout = append(stdout, LinkMessage(tc.Elf.Interpreter, filepath.Join(tmp, tc.Elf.Interpreter)))
	}
	libPath := filepath.Join(tmp, "lib64")
	for _, lib := range tc.Elf.Dependencies {
		stdout = append(stdout, LinkMessage(lib, filepath.Join(libPath, filepath.Base(lib))))
	}
	return
}
