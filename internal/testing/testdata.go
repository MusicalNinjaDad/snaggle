// Helpers related to testdata assets. Only suitable for integration tests (*_test) packages
// due to circular imports
package testing

//nolint:staticcheck
import (
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"

	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type binaryDetails struct {
	Description string // test run name
	ExpectedElf elf.Elf
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
			ExpectedElf: elf.Elf{
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
			ExpectedElf: elf.Elf{
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
			ExpectedElf: elf.Elf{
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
			ExpectedElf: elf.Elf{
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
	}
}
