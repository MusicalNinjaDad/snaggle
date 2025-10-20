package testdata

//nolint:staticcheck
import (
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type BinaryDetails struct {
	Name        string // test run name
	ExpectedElf elf.Elf
	Dynamic     bool
	Exe         bool
	Lib         bool
}

// Test Helper: provides a set of common binaries to use in table tests
func CommonBinaries(t *testing.T) []BinaryDetails {
	t.Helper()
	return []BinaryDetails{
		{
			Name: "PIE no dependencies",
			ExpectedElf: elf.Elf{
				Name:         "hello_pie",
				Path:         filepath.Join(Pwd(t), "../testdata/hello_pie"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: nil,
			},
			Dynamic: true,
			Exe:     true,
			Lib:     false,
		},
		{
			Name: "Static linked executable",
			ExpectedElf: elf.Elf{
				Name:         "hello_static",
				Path:         filepath.Join(Pwd(t), "../testdata/hello_static"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.EXE),
				Interpreter:  "",
				Dependencies: nil,
			},
			Dynamic: false,
			Exe:     true,
			Lib:     false,
		},
		{
			Name: "PIE 1 dependency",
			ExpectedElf: elf.Elf{
				Name:         "which",
				Path:         filepath.Join(Pwd(t), "../testdata/which"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: []string{P_libc},
			},
			Dynamic: true,
			Exe:     true,
			Lib:     false,
		},
		{
			Name: "PIE nested dependencies",
			ExpectedElf: elf.Elf{
				Name:         "id",
				Path:         filepath.Join(Pwd(t), "../testdata/id"),
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
