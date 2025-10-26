package elf

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestLdd_single_fedora(t *testing.T) {
	Assert := assert.New(t)
	expectedDependencies := []string{P_libc}
	dependencies, err := ldd(P_which, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLdd_single_ubuntu(t *testing.T) {
	Assert := assert.New(t)
	expectedDependencies := []string{P_libc}
	dependencies, err := ldd(P_which, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLdd_nested(t *testing.T) {
	Assert := assert.New(t)
	expectedDependencies := []string{P_libc, P_libpcre2_8, P_libselinux}
	dependencies, err := ldd(P_id, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLibpathcmp(t *testing.T) {
	fedora := "/lib64/libc.so.6"
	ubuntu := "/lib64/x86_64-linux-gnu/libc.so.6"
	assert.Zero(t, libpathcmp(fedora, ubuntu))
}

func TestSortByFilename(t *testing.T) {
	unsorted := []string{"/lib64/libpcre2-8.so.0", "/lib64/x86_64-linux-gnu/libselinux.so.1", "/lib64/x86_64-linux-gnu/libc.so.6"}
	sorted := []string{"/lib64/x86_64-linux-gnu/libc.so.6", "/lib64/libpcre2-8.so.0", "/lib64/x86_64-linux-gnu/libselinux.so.1"}
	assert.NotEqual(t, sorted, unsorted)
	slices.SortFunc(unsorted, libpathcmp)
	assert.Equal(t, sorted, unsorted)
}

// Test Helper: validates that a slice dependency paths are equal without breaking cross-platform
// portability (libs are potentially in different paths)
func AssertDependenciesEqual(t *testing.T, expected []string, actual []string) {
	t.Helper()
	if len(expected) == 0 && len(actual) == 0 {
		return // no dependencies
	}
	for idx, dep := range expected {
		assert.Zerof(t, libpathcmp(dep, actual[idx]), "dependency %v differs: %s != %s", idx, dep, actual[idx])
	}
}
