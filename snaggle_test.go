package snaggle

import (
	"slices"
	"testing"
)

func TestLddSingleBinary(t *testing.T) {
	expectedLibs := []string{"/lib64/libc.so.6", "/lib64/ld-linux-x86-64.so.2"}
	slices.Sort(expectedLibs)
	libs, err := LibPaths("/usr/sbin/which")
	slices.Sort(libs)
	if !slices.Equal(libs, expectedLibs) || err != nil {
		t.Error("Expected", expectedLibs, "got", libs, "with error", err)
	}
}