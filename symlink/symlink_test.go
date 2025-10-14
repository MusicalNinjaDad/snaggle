package symlink

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestNewSymlink(t *testing.T) {
	Assert := assert.New(t)
	source, _ := filepath.Abs("testdata/which")
	expectedLink := Symlink{Source: source, Target: "/usr/sbin/which", Err: nil}
	link := NewSymlink("testdata/which")
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
	source, _ := filepath.Abs("testdata/which")
	expectedChain := []Symlink{
		{Source: source, Target: "/usr/sbin/which", Err: nil},
		{Source: "/usr/sbin/which", Target: "", Err: nil},
	}
	chain := SymlinkChain("testdata/which")
	Assert.Equal(chain, expectedChain)
}
