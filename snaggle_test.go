package snaggle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLddSingleBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"/lib64/libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	slices.Sort(expectedLibs)
	libs, err := LibPaths("/usr/sbin/which")
	Assert.NoError(err)
	Assert.Equal(expectedLibs, libs)
}

func TestSymlinkedBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"/lib64/libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	slices.Sort(expectedLibs)
	libs, err := LibPaths("symlink/testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedLibs, libs)
}

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func workspaceTempDir(t *testing.T) string {
	t.Helper()
	tmpRoot := filepath.Join(pwd(t), ".tmp")
	err := os.Mkdir(tmpRoot, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatal("Failed to create ./.tmp. Error: ", err)
	}
	tmp, err := os.MkdirTemp(tmpRoot, t.Name())
	if err != nil {
		t.Fatal("Failed to create", tmp, "Error:", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmp) })
	return tmp
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
	tmp := workspaceTempDir(t)
	expectedFile := filepath.Join(tmp, pwd(t), "/testdata/which")
	path, err := CopyBinary("testdata/which", tmp)
	Assert.NoError(err)
	Assert.Equal(expectedFile, path)
	Assert.FileExists(expectedFile)
	Assert.True(SameFile(t, expectedFile, path))
}
