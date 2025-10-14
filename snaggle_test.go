package snaggle

import (
	"github.com/stretchr/testify/assert"
	"slices"
	"testing"
)

func TestLddSingleBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"/lib64/libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	slices.Sort(expectedLibs)
	libs, err := LibPaths("/usr/sbin/which")
	Assert.NoError(err)
	Assert.Equal(libs, expectedLibs)
}

func TestSymlinkedBinary(t *testing.T) {
	Assert := assert.New(t)
	expectedLibs := []string{"/lib64/libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	slices.Sort(expectedLibs)
	libs, err := LibPaths("test/assets/which")
	Assert.NoError(err)
	Assert.Equal(libs, expectedLibs)
}