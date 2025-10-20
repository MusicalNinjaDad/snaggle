package snaggle

import (
	"os"
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

func SameFile(t *testing.T, path1 string, path2 string) bool {
	t.Helper()
	file1, err1 := os.Stat(path1)
	file2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		t.Log("Errors stating:", path1, "-", err1, ";", path2, "-", err2)
		return false
	}
	return os.SameFile(file1, file2)
}

func TestCopyBinary(t *testing.T) {
	Assert := assert.New(t)
	tmp := WorkspaceTempDir(t)
	expectedFile := filepath.Join(tmp, Pwd(t), "/testdata/which")
	path, err := LinkFile("testdata/which", tmp)
	Assert.NoError(err)
	Assert.Equal(expectedFile, path)
	Assert.FileExists(expectedFile)
	Assert.True(SameFile(t, expectedFile, path))
}
