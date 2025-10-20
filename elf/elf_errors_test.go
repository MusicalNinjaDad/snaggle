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
	if err := noaccess.Chmod(0222); err != nil { // --w--w--w-
		t.Fatal(err)
	}
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

func static(t *testing.T) string {
	t.Helper()
	return filepath.Join(Pwd(t), "../testdata/hello_static")
}

func TestLddErrors(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		interpreter string
		errs        []error
		errorText   string
		invalidErrs []error
	}{
		{
			name:        "static binary",
			path:        static(t),
			interpreter: P_ld_linux,
			errs:        []error{ErrLdd},
			errorText:   "ldd failed to execute /lib64/ld-linux-x86-64.so.2 " + static(t) + ": ",
			invalidErrs: []error{ErrInvalidElf, ErrUnsupportedInterpreter},
		},
		{
			name:        "unsupported interpreter",
			path:        static(t),
			interpreter: "/lib64/evil_interpreter.so",
			errs:        []error{ErrUnsupportedInterpreter, ErrInvalidElf, errors.ErrUnsupported},
			errorText:   "invalid ELF file: unsupported operation (unsupported interpreter) '/lib64/evil_interpreter.so'",
			invalidErrs: []error{ErrLdd},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Assert := assert.New(t)
			dependencies, err := ldd(tc.path, tc.interpreter)
			Assert.Nil(dependencies)
			for _, e := range tc.errs {
				Assert.ErrorIs(err, e)
			}
			Assert.ErrorContains(err, tc.errorText)
			for _, e := range tc.invalidErrs {
				Assert.NotErrorIs(err, e)
			}
		})
	}
}
