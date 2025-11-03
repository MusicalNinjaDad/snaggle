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
			Assert := assert.New(t)

			path := tc.Elf.Path
			parsed, err := elf.New(path)

			Assert.NoError(err)
			Assert.Equal(tc.Exe, parsed.IsExe())
			Assert.Equal(tc.Lib, parsed.IsLib())
			Assert.Equal(tc.Dynamic, parsed.IsDyn())
			Assert.Nil(parsed.Diff(tc.Elf))
		})
	}
}

func Test(t *testing.T) {
	for t, tc := range TestCases(t) {
		Assert := assert.New(t)

		parsed, err := elf.New(tc.Src)

		Assert.NoError(err)
		Assert.Equal(tc.ExpectedBinary.Exe, parsed.IsExe())
		Assert.Equal(tc.ExpectedBinary.Lib, parsed.IsLib())
		Assert.Equal(tc.ExpectedBinary.Dynamic, parsed.IsDyn())
		Assert.Nil(parsed.Diff(tc.ExpectedBinary.Elf))

	}
}
