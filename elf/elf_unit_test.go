package elf

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestLdd_single_fedora(t *testing.T) {
	Assert := assert.New(t)
	which := filepath.Join(Pwd(t), "../testdata/which")
	expectedDependencies := []string{P_libc}
	dependencies, err := ldd(which, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLdd_single_ubuntu(t *testing.T) {
	Assert := assert.New(t)
	which := filepath.Join(Pwd(t), "../testdata/which")
	expectedDependencies := []string{P_libc}
	dependencies, err := ldd(which, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLdd_nested(t *testing.T) {
	Assert := assert.New(t)
	id := filepath.Join(Pwd(t), "../testdata/id")
	expectedDependencies := []string{P_libc, P_libpcre2_8, P_libselinux}
	dependencies, err := ldd(id, P_ld_linux)
	Assert.NoError(err)
	AssertDependenciesEqual(t, expectedDependencies, dependencies)
}

func TestLibpathcmp(t *testing.T) {
	fedora := "/lib64/libc.so.6"
	ubuntu := "/lib64/x86_64-linux-gnu/libc.so.6"
	assert.Zero(t, Libpathcmp(fedora, ubuntu))
}

func TestSortByFilename(t *testing.T) {
	unsorted := []string{"/lib64/libpcre2-8.so.0", "/lib64/x86_64-linux-gnu/libselinux.so.1", "/lib64/x86_64-linux-gnu/libc.so.6"}
	sorted := []string{"/lib64/x86_64-linux-gnu/libc.so.6", "/lib64/libpcre2-8.so.0", "/lib64/x86_64-linux-gnu/libselinux.so.1"}
	assert.NotEqual(t, sorted, unsorted)
	slices.SortFunc(unsorted, Libpathcmp)
	assert.Equal(t, sorted, unsorted)
}
