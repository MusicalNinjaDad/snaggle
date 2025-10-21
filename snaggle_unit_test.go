package snaggle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestLinkTree(t *testing.T) {
	Assert := assert.New(t)
	tmp := WorkspaceTempDir(t)
	expectedFile := filepath.Join(tmp, P_which)
	path, err := linkTree(P_which, tmp)
	Assert.NoError(err)
	Assert.Equal(expectedFile, path)
	Assert.FileExists(expectedFile)
	AssertSameInode(t, expectedFile, path)
}
