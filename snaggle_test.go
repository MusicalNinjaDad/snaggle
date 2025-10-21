package snaggle_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
)

func TestWhich(t *testing.T) {
	Assert := assert.New(t)
	path := P_which
	tmp := WorkspaceTempDir(t)
	expectedBin := filepath.Join(tmp, "bin", filepath.Base(P_which))
	expectedLib := filepath.Join(tmp, "lib", filepath.Base(P_libc))
	err := Snaggle(path)
	Assert.NoError(err)
	AssertSameFile(t, P_which, expectedBin)
	AssertSameFile(t, P_libc, expectedLib)
}
