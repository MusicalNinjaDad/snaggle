package elf_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestCommonBinaries(t *testing.T) {
	tests := []struct {
		name        string // test run name
		path        string
		expectedElf elf.Elf
		dynamic     bool
		exe         bool
		lib         bool
	}{
		{
			name: "PIE no dependencies",
			expectedElf: elf.Elf{
				Name:         "hello_pie",
				Path:         filepath.Join(Pwd(t), "../testdata/hello_pie"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: nil,
			},
			dynamic: true,
			exe:     true,
			lib:     false,
		},
		{
			name: "Static linked executable",
			expectedElf: elf.Elf{
				Name:         "hello_static",
				Path:         filepath.Join(Pwd(t), "../testdata/hello_static"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.EXE),
				Interpreter:  "",
				Dependencies: nil,
			},
			dynamic: false,
			exe:     true,
			lib:     false,
		},
		{
			name: "PIE 1 dependency",
			expectedElf: elf.Elf{
				Name:         "which",
				Path:         filepath.Join(Pwd(t), "../testdata/which"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: []string{P_libc},
			},
			dynamic: true,
			exe:     true,
			lib:     false,
		},
		{
			name: "PIE nested dependencies",
			expectedElf: elf.Elf{
				Name:         "id",
				Path:         filepath.Join(Pwd(t), "../testdata/id"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  P_ld_linux,
				Dependencies: []string{P_libc, P_libpcre2_8, P_libselinux},
			},
			dynamic: true,
			exe:     true,
			lib:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.expectedElf.Path
			if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("Skipping %s as %s not present", tt.name, path)
				return
			}

			Assert := assert.New(t)
			parsed, err := elf.New(path)
			Assert.NoError(err)
			Assert.Equal(tt.exe, parsed.IsExe())
			Assert.Equal(tt.lib, parsed.IsLib())
			Assert.Equal(tt.dynamic, parsed.IsDyn())
			Assert.Nil(parsed.Diff(tt.expectedElf))
		})
	}
}
