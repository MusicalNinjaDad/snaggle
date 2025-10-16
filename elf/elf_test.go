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
		Path: filepath.Join(pwd(t), "../testdata/which"),
	}
	whichElf, err := elf.New("../testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedElf, whichElf)
}
