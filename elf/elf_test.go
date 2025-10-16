package elf_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/stretchr/testify/assert"
)

type Elf = elf.Elf

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func TestWhich(t *testing.T) {
	Assert := assert.New(t)
	expectedElf := Elf{
		Name:        "which",
		Path:        filepath.Join(pwd(t), "../testdata/which"),
		Class:       elf.EI_CLASS(elf.ELF64),
		Interpreter: "/lib64/ld-linux-x86-64.so.2",
		Type:        elf.Type(elf.BIN),
	}
	whichElf, err := elf.New("../testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedElf, whichElf)
}
