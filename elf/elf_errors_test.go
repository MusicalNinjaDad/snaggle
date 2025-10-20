package elf

import (
	"debug/elf"
	"errors"
	"io/fs"
	"os"
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
	Assert.NotErrorIs(err, ErrInvalidElf)
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
	Assert.ErrorIs(err, ErrLdd)
	Assert.ErrorContains(err, "ldd failed to execute /lib64/ld-linux-x86-64.so.2 "+static+": ")
	Assert.Nil(dependencies)
}

func TestLdd_unsupported_interpreter(t *testing.T) {
	Assert := assert.New(t)
	static := filepath.Join(Pwd(t), "../testdata/hello_static")
	interpreter := "/lib64/evil_interpreter.so"
	dependencies, err := ldd(static, interpreter)
	Assert.ErrorIs(err, ErrUnsupportedInterpreter)
	Assert.ErrorIs(err, ErrInvalidElf)
	Assert.ErrorIs(err, errors.ErrUnsupported)
	Assert.ErrorContains(err, "invalid ELF file: unsupported operation (unsupported interpreter) '"+interpreter+"'")
	Assert.Nil(dependencies)
}

func TestIO_Error(t *testing.T) {
	Assert := assert.New(t)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "noaccess")
	noaccess, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(noaccess.Name())
	noaccess.Chmod(0222) // --w--w--w-
	bad, err := New(noaccess.Name())
	var errelf *ErrElf
	Assert.Equal("noaccess", bad.Name)
	Assert.Equal(path, bad.Path)
	Assert.ErrorIs(err, fs.ErrPermission)
	Assert.NotErrorIs(err, ErrInvalidElf)
	Assert.ErrorAs(err, &errelf)
	errpath := err.(*ErrElf).Path()
	Assert.Equal(path, errpath)
	Assert.ErrorContains(err, "error(s) parsing "+path+":")
	Assert.ErrorContains(err, "permission denied")
}
