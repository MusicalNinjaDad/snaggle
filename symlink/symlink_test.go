package symlink_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/MusicalNinjaDad/snaggle/symlink"
	"github.com/stretchr/testify/assert"
)

type Symlink = symlink.Symlink

func pwd(t *testing.T) string {
	t.Helper()
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal("Failed to get pwd. Error:", err)
	}
	return pwd
}

func TestNewSymlink(t *testing.T) {
	Assert := assert.New(t)
	sourcelink := filepath.Join(pwd(t), "testdata/which")
	expectedLink := Symlink{Source: sourcelink, Target: "../../testdata/which", Err: nil}
	link := symlink.New("testdata/which")
	Assert.Equal(expectedLink, link)
}

func TestNewSymlinkNotALink(t *testing.T) {
	Assert := assert.New(t)
	link := symlink.New("../testdata/which")
	Assert.True(link.IsNotLink())
}

func TestSymlinkChain(t *testing.T) {
	Assert := assert.New(t)
	sourcelink := filepath.Join(pwd(t), "testdata/which")
	ourwhich := filepath.Join(pwd(t), "../testdata/which")
	expectedChain := []Symlink{
		{Source: sourcelink, Target: "../../testdata/which", Err: nil},
		{Source: ourwhich, Target: "", Err: nil},
	}
	chain, err := symlink.Chain("testdata/which")
	Assert.NoError(err)
	Assert.Equal(expectedChain, chain)
}
