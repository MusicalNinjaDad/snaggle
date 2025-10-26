// Helpers and values which can be used in any package
package internal

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/ameghdadian/x/iter"

	"github.com/stretchr/testify/assert"
)

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
func WorkspaceTempDir(t testing.TB) string {
	t.Helper()
	tmpRoot := TestdataPath(".tmp")
	err := os.Mkdir(tmpRoot, os.ModePerm)
	if err != nil && !errors.Is(err, os.ErrExist) {
		t.Fatal("Failed to create ./.tmp. Error: ", err)
	}
	tmp, err := os.MkdirTemp(tmpRoot, strings.ReplaceAll(t.Name(), "/", "_"))
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

// Paths to our test binaries
var (
	P_hello_pie    = TestdataPath("hello_pie")
	P_hello_static = TestdataPath("hello_static")
	P_which        = TestdataPath("which")
	P_id           = TestdataPath("id")
	P_ldd          = TestdataPath("ldd")
	P_ctypes_so    = TestdataPath("_ctypes_test.cpython-314-x86_64-linux-gnu.so")
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

// StripLines converts a multiline string into a []string without line termination or indentation.
func StripLines(multiline string) []string {
	return slices.Collect(iter.Map(strings.Lines(multiline), strings.TrimSpace))
}

// Build a message about a link attempt for stdout
func LinkMessage(from string, to string) string {
	return from + " -> " + to
}
