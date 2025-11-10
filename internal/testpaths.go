package internal

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Paths to our test binaries
var (
	P_hello_pie     = TestdataPath("hello_pie")
	P_hello_static  = TestdataPath("hello_static")
	P_which         = TestdataPath("which")
	P_id            = TestdataPath("id")
	P_ldd           = TestdataPath("ldd")
	P_ctypes_so     = TestdataPath("_ctypes_test.cpython-314-x86_64-linux-gnu.so")
	P_build_sh      = TestdataPath("hello/build.sh")
	P_hello_dynamic = TestdataPath("hello/hello")
	P_hello_go      = TestdataPath("hello/hello.go")
	P_symlinked_id  = TestdataPath("symlink/id2")
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

// Constructs a TempDir under `./.tmp`
//
// This is (almost) guaranteed to be on the same filesystem as `./internal/testdata` and therefore
// allow for valid hardlinks.
func WorkspaceTempDir(t testing.TB) string {
	t.Helper()
	tmpRoot := TestdataPath("../../.tmp")
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

// Constructs the correct absolute value of path, where path is relative to ./internal/testdata
func TestdataPath(path string) string {
	_, thisfile, _, _ := runtime.Caller(0)
	internalDir := filepath.Dir(thisfile)
	return filepath.Join(internalDir, "testdata", path)
}
