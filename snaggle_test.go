package snaggle_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle"
	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			var stdout bytes.Buffer
			log.SetOutput(&stdout)
			t.Cleanup(func() { log.SetOutput(os.Stdout) })

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			binPath := filepath.Join(tmp, "bin", filepath.Base(tc.ExpectedElf.Name))
			var libCopies []string
			for _, lib := range tc.ExpectedElf.Dependencies {
				libCopies = append(libCopies,
					filepath.Join(tmp, "lib64", filepath.Base(lib)),
				)
			}
			msg := tc.ExpectedElf.Path + " -> " + binPath

			err := snaggle.Snaggle(tc.ExpectedElf.Path, tmp)
			Assert.NoError(err)
			AssertSameInode(t, tc.ExpectedElf.Path, binPath)
			for idx, copy := range libCopies {
				original := tc.ExpectedElf.Dependencies[idx]
				same := SameFile(original, copy)
				assert.Truef(t, same, "%s & %s are different files", original, copy)
			}

			Assert.Contains(stdout.String(), msg)
		})
	}
}
