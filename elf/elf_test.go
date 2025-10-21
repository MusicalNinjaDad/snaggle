package elf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/elf"

	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			path := tc.ExpectedElf.Path
			Assert := assert.New(t)
			parsed, err := elf.New(path)
			Assert.NoError(err)
			Assert.Equal(tc.Exe, parsed.IsExe())
			Assert.Equal(tc.Lib, parsed.IsLib())
			Assert.Equal(tc.Dynamic, parsed.IsDyn())
			Assert.Nil(parsed.Diff(tc.ExpectedElf))
		})
	}
}
