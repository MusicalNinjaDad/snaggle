package internal

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// If both paths are absolute: compares only the filename, otherwise compares the entire path.
func Libpathcmp(path1 string, path2 string) int {
	if filepath.IsAbs(path1) && filepath.IsAbs(path2) {
		return strings.Compare(filepath.Base(path1), filepath.Base(path2))
	}
	return strings.Compare(path1, path2)
}

// Regex to check if this is a 64-bit version of `ld-linux*.so`, matches /lib64(/more/directories)/ld-linux*.so(.*)
var Ld_linux_64_RE = regexp.MustCompile(`^\/lib64(?:\/.+|)\/ld-linux.*\.so(?:\..+|)$`)

// Test helper: returns current working directory, will `FailNow` on error
func Pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

// Construct a TempDir under `./.tmp`
//
// This is (almost) guaranteed to be on the same filesystem as `./testdata` and therefore
// allow for valid hardlinks.
func WorkspaceTempDir(t *testing.T) string {
	t.Helper()
	tmpRoot := filepath.Join(Pwd(t), ".tmp")
	err := os.Mkdir(tmpRoot, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatal("Failed to create ./.tmp. Error: ", err)
	}
	tmp, err := os.MkdirTemp(tmpRoot, t.Name())
	if err != nil {
		t.Fatal("Failed to create", tmp, "Error:", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmp) })
	return tmp
}

// Test Helper: validates that a slice dependency paths are equal without breaking cross-platform
// portability (libs are potentially in different paths)
func AssertDependenciesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	for idx, dep := range expected {
		assert.Zerof(t, Libpathcmp(dep, actual[idx]), "dependency %v differs: %s != %s", idx, dep, actual[idx])
	}
}

// Test Helper: validates that two paths point to the same inode
func AssertSameFile(t *testing.T, path1 string, path2 string) {
	t.Helper()
	file1, err1 := os.Stat(path1)
	file2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		t.Fatal("Errors stating:", path1, "-", err1, ";", path2, "-", err2)
	}
	assert.Truef(t, os.SameFile(file1, file2), "%s & %s are different files", file1, file2)
}

// Paths to common libraries
const (
	P_ld_linux   = "/lib64/ld-linux-x86-64.so.2"
	P_libc       = "/lib64/libc.so.6"
	P_libpcre2_8 = "/lib64/libpcre2-8.so.0"
	P_libselinux = "/lib64/libselinux.so.1"
)
