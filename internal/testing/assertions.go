package testing

import (
	"io/fs"
	"maps"
	"path/filepath"
	"slices"
	"strings"
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

func (a *Asserter) LinkedFile(path1 string, path2 string) {
	a.t.Helper()
	same, err := SameInode(path1, path2)
	a.Testify.NoError(err)
	a.Testify.Truef(same, "%s & %s are different files", path1, path2)
}

// Assert:
//  1. Actual contains all expected lines, in correct order (ignoring prefix "link" / "copy")
//  1. All lines are prefixed with either "link" or "copy"
//  1. All lines referencing paths in mustBeLinked are prefixed with "link"
func (a *Asserter) Stdout(expected []string, actual []string, mustBeLinked ...string) {
	a.t.Helper()

	linked := make(map[string]bool, len(expected))

	stripped := make([]string, 0, len(actual))
	for n, line := range actual {
		if strippedline, ok := strings.CutPrefix(line, "copy "); ok {
			stripped = append(stripped, strippedline)
			linked[strings.Fields(strippedline)[0]] = false
			continue
		}
		if strippedline, ok := strings.CutPrefix(line, "link "); ok {
			stripped = append(stripped, strippedline)
			linked[strings.Fields(strippedline)[0]] = true
			continue
		}
		a.t.Errorf("Line %v does not start with `copy` or `line`: %s", n+1, line)
	}

	// TODO with #84 - assert Equal (ordering guaranteed with --verbose)
	a.Testify.ElementsMatch(expected, stripped)

	for _, filename := range mustBeLinked {
		a.Testify.Truef(linked[filename], "%s was reported as copied", filename)
	}

}
