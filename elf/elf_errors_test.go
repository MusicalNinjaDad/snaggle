package elf

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileNotFound(t *testing.T) {
	Assert := assert.New(t)
	const path = "/bad/path"
	bad, err := New(path)
	var errelf *ErrElf
	Assert.Equal("path", bad.Name)
	Assert.ErrorIs(err, fs.ErrNotExist)
	Assert.ErrorAs(err, &errelf)
	Assert.Equal(path, bad.Path)
}
