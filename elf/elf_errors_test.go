package elf

import (
	"debug/elf"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestFileNotFound(t *testing.T) {
	Assert := assert.New(t)
	const path = "/bad/path"
	bad, err := New(path)
	var errelf *ErrElf
	Assert.Equal("path", bad.Name)
	Assert.Equal(path, bad.Path)
	Assert.ErrorIs(err, fs.ErrNotExist)
	Assert.ErrorAs(err, &errelf)
	errpath := err.(*ErrElf).Path()
	Assert.Equal(path, errpath)
	Assert.ErrorContains(err, "error(s) parsing /bad/path:")
	Assert.ErrorContains(err, "no such file or directory")
}

func TestNotElf(t *testing.T) {
	Assert := assert.New(t)
	path := filepath.Join(Pwd(t), "../testdata/ldd")
	bad, err := New(path)
	var errelf *ErrElf
	var errformat *elf.FormatError
	Assert.Equal("ldd", bad.Name)
	Assert.Equal(path, bad.Path)
	Assert.ErrorAs(err, &errformat)
	Assert.ErrorIs(err, ErrInvalidElf)
	Assert.ErrorAs(err, &errelf)
	errpath := err.(*ErrElf).Path()
	Assert.Equal(path, errpath)
	Assert.ErrorContains(err, "error(s) parsing "+path+":")
	Assert.ErrorContains(err, "invalid ELF file")
}

func TestLdd_static(t *testing.T) {
	Assert := assert.New(t)
	static := filepath.Join(Pwd(t), "../testdata/hello_static")
	dependencies, err := ldd(static, P_ld_linux)
	Assert.ErrorIs(err, ErrElfLdd)
	Assert.ErrorContains(err, "ldd failed to execute /lib64/ld-linux-x86-64.so.2 /workspaces/snaggle/testdata/hello_static: ")
	Assert.Nil(dependencies)
}
