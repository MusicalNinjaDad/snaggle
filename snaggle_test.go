package snaggle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestLddSingleBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	libs, err := LibPaths("testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedLibs, libs)
}

func TestSymlinkedBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	libs, err := LibPaths("testdata/which2")
	Assert.NoError(err)
	Assert.Equal(expectedLibs, libs)
}

func TestCopyBinary(t *testing.T) {
	Assert := assert.New(t)
	tmp := WorkspaceTempDir(t)
	expectedFile := filepath.Join(tmp, Pwd(t), "/testdata/which")
	path, err := LinkFile("testdata/which", tmp)
	Assert.NoError(err)
	Assert.Equal(expectedFile, path)
	Assert.FileExists(expectedFile)
	AssertSameFile(t, expectedFile, path)
}
