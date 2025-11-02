package testing

import (
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"

	//nolint:staticcheck
	. "github.com/MusicalNinjaDad/snaggle/internal" //lint:ignore ST1001 test helpers
)

type Asserter struct {
	Testify *assert.Assertions
	t       *testing.T
}

func Assert(t *testing.T) *Asserter {
	Testify := assert.New(t)
	return &Asserter{Testify, t}
}

// Assert contents of dir match map[reference_file]expected_file.
//
//  1. Asserts ALL expected files are present
//  1. Asserts all expected files are IDENTICAL to the reference files
//  1. Assert NO OTHER files are present in dir
func (a *Asserter) DirectoryContents(ExpectedContents map[string]string, dir string) {
	a.t.Helper()

	contents := make([]string, 0, len(ExpectedContents))
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			contents = append(contents, path)
		}
		return err // lazy - append may have already happened but not important.
	})
	if !a.Testify.NoError(err) {
		a.t.Fatal("Error listing directory contents")
	}

	// expected is coming from a map, so has non-deterministic ordering
	a.Testify.ElementsMatch(slices.Collect(maps.Values(ExpectedContents)), contents)

	for original, copy := range ExpectedContents {
		a.Testify.Truef(SameFile(original, copy), "%s & %s are different files", original, copy)
	}

}
