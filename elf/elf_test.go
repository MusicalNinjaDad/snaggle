package elf_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle/elf"

	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func Test(t *testing.T) {
	for name, details := range GoodElfs {
		t.Run(name, func(t *testing.T) {
			Assert := assert.New(t)

			parsed, err := elf.New(details.Elf.Path)

			Assert.NoError(err)
			Assert.Equal(details.Exe, parsed.IsExe())
			Assert.Equal(details.Lib, parsed.IsLib())
			Assert.Equal(details.Dynamic, parsed.IsDyn())
			Assert.Nil(parsed.Diff(details.Elf))
		})
	}
}
