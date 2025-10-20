package snaggle

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal/test"
)

func TestCopyBinary(t *testing.T) {
	Assert := assert.New(t)
	tmp := WorkspaceTempDir(t)
	expectedFile := filepath.Join(tmp, Pwd(t), "/testdata/which")
	path, err := linkFile("testdata/which", tmp)
	Assert.NoError(err)
	Assert.Equal(expectedFile, path)
	Assert.FileExists(expectedFile)
	AssertSameFile(t, expectedFile, path)
}
