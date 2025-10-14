package snaggle

import (
	"github.com/stretchr/testify/assert"
	"path/filepath"
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

func TestSymlinkTree(t *testing.T) {
	Assert := assert.New(t)
	expectedTree := make(map[string]string)
	link, _ := filepath.Abs("test/assets/which")
	expectedTree[link] = "/usr/sbin/which"
	tree, err := SymlinkTree("test/assets/which")
	Assert.NoError(err)
	Assert.Equal(tree, expectedTree)
}
