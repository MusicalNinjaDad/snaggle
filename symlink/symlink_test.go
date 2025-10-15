package symlink_test

import (
	"testing"

	"github.com/MusicalNinjaDad/snaggle/symlink"
	"github.com/stretchr/testify/assert"
)

type Symlink = symlink.Symlink

func TestNewSymlink(t *testing.T) {
	Assert := assert.New(t)
	expectedLink := Symlink{Source: "testdata/which", Target: "/usr/sbin/which", Err: nil}
	link := symlink.New("testdata/which")
	Assert.Equal(expectedLink, link)
}

func TestNewSymlinkNotALink(t *testing.T) {
	Assert := assert.New(t)
	link := symlink.New("/usr/sbin/which")
	Assert.True(link.IsNotLink())
}

func TestSymlinkChain(t *testing.T) {
	Assert := assert.New(t)
	expectedChain := []Symlink{
		{Source: "testdata/which", Target: "/usr/sbin/which", Err: nil},
		{Source: "/usr/sbin/which", Target: "", Err: nil},
	}
	chain, err := symlink.Chain("testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedChain, chain)
}
