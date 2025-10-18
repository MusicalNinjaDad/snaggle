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

func TestLibselinux(t *testing.T) {
	expectedElf := elf.Elf{
		Name:         "libselinux.so.1",
		Path:         "/lib64/libselinux.so.1",
		Class:        elf.EI_CLASS(elf.ELF64),
		Type:         elf.Type(elf.DYN),
		Interpreter:  "",
		Dependencies: []string{"libc.so.6", "libpcre2-8.so.0"},
	}
	Assert := assert.New(t)
	parsed, err := elf.New(expectedElf.Path)
	Assert.NoError(err)
	Assert.False(parsed.IsExe(), "IsExe()")
	Assert.True(parsed.IsLib(), "IsLib()")
	Assert.True(parsed.IsDyn(), "IsDyn()")
	// Assert.Equal(expectedElf, parsed)
}

func TestCommonBinaries(t *testing.T) {
	tests := []struct {
		name        string // test run name
		path        string
		expectedElf elf.Elf
		dynamic     bool
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
		},
		{
			name: "PIE 1 dependency",
			expectedElf: elf.Elf{
				Name:         "which",
				Path:         filepath.Join(pwd(t), "../testdata/which"),
				Class:        elf.EI_CLASS(elf.ELF64),
				Type:         elf.Type(elf.PIE),
				Interpreter:  "/lib64/ld-linux-x86-64.so.2",
				Dependencies: []string{"libc.so.6"},
			},
			dynamic: true,
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
				Dependencies: []string{"libc.so.6", "libselinux.so.1"},
			},
			dynamic: true,
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
			Assert.True(parsed.IsExe(), "IsExe()")
			Assert.Equal(tt.dynamic, parsed.IsDyn())
			Assert.Equal(tt.expectedElf, parsed)
		})
	}
}
