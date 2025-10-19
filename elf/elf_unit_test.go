package elf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func TestLdd_single(t *testing.T) {
	Assert := assert.New(t)
	which := filepath.Join(pwd(t), "../testdata/which")
	expectedDependencies := []string{"/lib64/libc.so.6"}
	dependencies, err := ldd(which)
	Assert.NoError(err)
	Assert.ElementsMatch(expectedDependencies, dependencies)
}

func TestLdd_nested(t *testing.T) {
	Assert := assert.New(t)
	id := filepath.Join(pwd(t), "../testdata/id")
	expectedDependencies := []string{"/lib64/libc.so.6", "/lib64/libpcre2-8.so.0", "/lib64/libselinux.so.1"}
	dependencies, err := ldd(id)
	Assert.NoError(err)
	Assert.ElementsMatch(expectedDependencies, dependencies)
}

func TestLdd_static(t *testing.T) {
	Assert := assert.New(t)
	id := filepath.Join(pwd(t), "../testdata/hello_static")
	dependencies, err := ldd(id)
	Assert.ErrorIs(err, ErrElf)
	Assert.ErrorIs(err, ErrElfLdd)
	Assert.Nil(dependencies)
}
