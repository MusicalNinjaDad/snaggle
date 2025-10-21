package snaggle_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle"
	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestWhich(t *testing.T) {
	Assert := assert.New(t)
	path := P_which
	tmp := WorkspaceTempDir(t)
	expectedBin := filepath.Join(tmp, "bin", filepath.Base(P_which))
	expectedLib := filepath.Join(tmp, "lib64", filepath.Base(P_libc))
	err := Snaggle(path, tmp)
	Assert.NoError(err)
	AssertSameInode(t, P_which, expectedBin)
	AssertSameInode(t, P_libc, expectedLib)
}
