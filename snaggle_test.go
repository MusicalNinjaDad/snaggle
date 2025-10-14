package snaggle

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestNewSymlink(t *testing.T) {
	Assert := assert.New(t)
	source, _ := filepath.Abs("test/assets/which")
	expectedLink := Symlink{Source: source, Target: "/usr/sbin/which", Err: nil}
	link := NewSymlink("test/assets/which")
	Assert.Equal(link, expectedLink)
}

func TestNewSymlinkNotALink(t *testing.T) {
	Assert := assert.New(t)
	expectedLink := Symlink{Source: "/usr/sbin/which", Target: "", Err: nil}
	link := NewSymlink("/usr/sbin/which")
	Assert.Equal(link, expectedLink)
}

func TestSymlinkChain(t *testing.T) {
	Assert := assert.New(t)
	source, _ := filepath.Abs("test/assets/which")
	expectedChain := []Symlink{
		{Source: source, Target: "/usr/sbin/which", Err: nil},
		{Source: "/usr/sbin/which", Target: "", Err: nil},
	}
	chain := SymlinkChain("test/assets/which")
	Assert.Equal(chain, expectedChain)
}
