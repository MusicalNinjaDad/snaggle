package elf_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/stretchr/testify/assert"
)

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func TestCommonBinaries(t *testing.T) {
	ldso := elf.Elf{
		Name:         "ld-linux-x86-64.so.2",
		Path:         "/lib64/ld-linux-x86-64.so.2",
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "",
		Dependencies: nil,
	}
	libc := elf.Elf{
		Name:         "libc.so.6",
		Path:         "/lib64/libc.so.6",
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "/lib64/ld-linux-x86-64.so.2",
		Dependencies: []elf.Elf{ldso},
	}
	libpcre2_8 := elf.Elf{
		Name:         "libpcre2-8.so.0",
		Path:         "/lib64/libpcre2-8.so.0.14.0",
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "",
		Dependencies: []elf.Elf{libc},
	}
	libselinux := elf.Elf{
		Name:         "libselinux.so.1",
		Path:         "/lib64/libselinux.so.1",
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "",
		Dependencies: []elf.Elf{libc, libpcre2_8},
	}
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
				Path:         filepath.Join(pwd(t), "../testdata/hello_pie"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  "/lib64/ld-linux-x86-64.so.2",
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
				Path:         filepath.Join(pwd(t), "../testdata/hello_static"),
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
				Path:         filepath.Join(pwd(t), "../testdata/which"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  "/lib64/ld-linux-x86-64.so.2",
				Dependencies: []elf.Elf{libc},
			},
			dynamic: true,
			exe:     true,
			lib:     false,
		},
		{
			name: "PIE nested dependencies",
			expectedElf: elf.Elf{
				Name:        "id",
				Path:        filepath.Join(pwd(t), "../testdata/id"),
				Class:       elf.EI_CLASS(elf.ELF64),
				Type:        elf.Type(elf.PIE),
				Interpreter: "/lib64/ld-linux-x86-64.so.2",
				// ldd lists "libpcre2-8.so.0", which is requested by "libselinux.so.1"
				Dependencies: []elf.Elf{libc, libselinux},
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
