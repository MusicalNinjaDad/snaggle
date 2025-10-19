package internal

import (
	"os"
	"path/filepath"
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

// Test helper: returns current working directory, will `FailNow` on error
func Pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

// Test Helper: validates that a slice dependency paths are equal without breaking cross-platform
// portability (libs are potentially in different paths)
func AssertDependenciesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	for idx, dep := range expected {
		assert.Zero(t, Libpathcmp(dep, actual[idx]))
	}
}

// Paths to common libraries
const (
	P_ld_linux   = "/lib64/ld-linux-x86-64.so.2"
	P_libc       = "/lib64/libc.so.6"
	P_libpcre2_8 = "/lib64/libpcre2-8.so.0"
	P_libselinux = "/lib64/libselinux.so.1"
)
