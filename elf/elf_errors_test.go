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

func readOnlyFile(t *testing.T) *os.File {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "noaccess")
	noaccess, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	noaccess.Chmod(0222) // --w--w--w-
	return noaccess
}

func TestFileErrors(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		path      string
		err       error
		errorText string
	}{
		{
			name:      "DirectoryNotFound",
			filename:  "nonexistant",
			path:      filepath.Join(Pwd(t), "../testdata/nothere/nonexistant"),
			err:       fs.ErrNotExist,
			errorText: "no such file or directory",
		},
		{
			name:      "Access Denied",
			filename:  "noaccess",
			path:      readOnlyFile(t).Name(),
			err:       fs.ErrPermission,
			errorText: "permission denied",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Assert := assert.New(t)
			var errelf *ErrElf
			parsed, err := New(tc.path)
			// We should have basic info in ALL cases
			Assert.Equal(parsed.Name, tc.filename)
			Assert.Equal(parsed.Path, tc.path)

			Assert.ErrorIs(err, tc.err)
			Assert.ErrorContains(err, tc.errorText)

			Assert.ErrorAs(err, &errelf)
			Assert.Equal(tc.path, errelf.Path())
			Assert.ErrorContains(err, "error(s) parsing "+tc.path+":")

			Assert.NotErrorIs(err, ErrInvalidElf)
		})
	}
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
