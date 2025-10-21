// Helpers and values which can be used in any package
package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
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

// Test helper: construct the correct absolute value of path, where path is relative to ./internal/testdata
func TestdataPath(path string) string {
	_, thisfile, _, _ := runtime.Caller(0)
	internalDir := filepath.Dir(thisfile)
	return filepath.Join(internalDir, "testdata", path)
}

// Test helper: constructs a TempDir under `./internal/testdata/.tmp`
//
// This is (almost) guaranteed to be on the same filesystem as `./internal/testdata` and therefore
// allow for valid hardlinks.
func WorkspaceTempDir(t *testing.T) string {
	t.Helper()
	tmpRoot := TestdataPath(".tmp")
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
	if len(expected) == 0 && len(actual) == 0 {
		return // no dependencies
	}
	for idx, dep := range expected {
		assert.Zerof(t, Libpathcmp(dep, actual[idx]), "dependency %v differs: %s != %s", idx, dep, actual[idx])
	}
}

// Are two paths refering to identical inodes?
func sameInode(path1 string, path2 string) (bool, error) {
	file1, err1 := os.Stat(path1)
	file2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		err := fmt.Errorf("errors stating: %s - %w; %s - %w", path1, err1, path2, err2)
		return false, err
	}
	return os.SameFile(file1, file2), nil
}

// Test Helper: validates that two paths point to the same inode
func AssertSameInode(t *testing.T, path1 string, path2 string) {
	t.Helper()
	same, err := sameInode(path1, path2)
	assert.NoError(t, err)
	assert.Truef(t, same, "%s & %s are different files", path1, path2)
}

// Do both files have same SHA256?
func sameHash(path1 string, path2 string) (bool, error) {
	sha1, err1 := hashFile(path1)
	sha2, err2 := hashFile(path2)
	return slices.Equal(sha1, sha2), errors.Join(err1, err2)
}

func hashFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err,
			file.Close(),
		)
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// Test Helper: validates that two files are identical
func AssertSameFile(t *testing.T, path1 string, path2 string) {
	t.Helper()
	same, err := sameInode(path1, path2)
	assert.NoError(t, err)
	if !same {
		same, err = sameHash(path1, path2)
		assert.NoError(t, err)
		assert.Truef(t, same, "%s & %s are different files", path1, path2)
	}
}

// Paths to common libraries
var (
	P_ld_linux   = findLib("ld-linux-x86-64.so.2")
	P_libc       = findLib("libc.so.6")
	P_libpcre2_8 = findLib("libpcre2-8.so.0")
	P_libselinux = findLib("libselinux.so.1")
)

// searches /lib* & /usr/lib* to find filename.
func findLib(filename string) (path string) {
	searchPaths := []string{"/lib*/*/", "/usr/lib*/*/", "/lib*/", "/usr/lib*/"} // ld.so-like ordering
	for _, dir := range searchPaths {
		matches, _ := filepath.Glob(dir + filename) // only possible returned error is ErrBadPattern
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

// Paths to our test binaries
var (
	P_hello_pie    = TestdataPath("hello_pie")
	P_hello_static = TestdataPath("hello_static")
	P_which        = TestdataPath("which")
	P_id           = TestdataPath("id")
	P_ldd          = TestdataPath("ldd")
)

// Test helper: Provides a temporary file which exists but has no read permission
func PermissionDenied(t *testing.T, filename string) *os.File {
	t.Helper()
	tmp := t.TempDir()
	path := filepath.Join(tmp, filename)
	noaccess, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := noaccess.Chmod(0222); err != nil { // --w--w--w-
		t.Fatal(err)
	}
	return noaccess
}
