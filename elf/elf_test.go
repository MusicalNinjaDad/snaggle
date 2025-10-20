package elf_test

import (
	"errors"
	"io/fs"
	"os"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal/testdata"
)

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			path := tt.ExpectedElf.Path
			if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("Skipping %s as %s not present", tt.Name, path)
				return
			}

			Assert := assert.New(t)
			parsed, err := elf.New(path)
			Assert.NoError(err)
			Assert.Equal(tt.Exe, parsed.IsExe())
			Assert.Equal(tt.Lib, parsed.IsLib())
			Assert.Equal(tt.Dynamic, parsed.IsDyn())
			Assert.Nil(parsed.Diff(tt.ExpectedElf))
		})
	}
}
