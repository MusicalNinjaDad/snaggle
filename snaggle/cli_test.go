package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/ameghdadian/x/iter"
	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	buildTmp := t.TempDir()
	_, thisfile, _, _ := runtime.Caller(0)
	srcDir := filepath.Dir(thisfile)
	build := exec.Command("go", "build", "-o", buildTmp, srcDir)
	if err := build.Run(); err != nil {
		t.Fatalf("cannot %s: %v", build.Args, err)
	}
	snaggleBin := filepath.Join(buildTmp, "snaggle")

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)
			snaggle := exec.Command(snaggleBin, tc.ExpectedElf.Path, tmp)

			binPath := filepath.Join(tmp, "bin", filepath.Base(tc.ExpectedElf.Name))
			expectedOut := make([]string, 0, 1+len(tc.ExpectedElf.Dependencies))
			expectedOut = append(expectedOut, tc.ExpectedElf.Path+" -> "+binPath)
			var libCopies []string
			for _, lib := range tc.ExpectedElf.Dependencies {
				copy := filepath.Join(tmp, "lib64", filepath.Base(lib))
				libCopies = append(libCopies, copy)
				expectedOut = append(expectedOut, lib+" -> "+copy)
			}

			stdout, err := snaggle.Output()
			Assert.NoError(err)
			AssertSameInode(t, tc.ExpectedElf.Path, binPath)
			for idx, copy := range libCopies {
				original := tc.ExpectedElf.Dependencies[idx]
				same := SameFile(original, copy)
				assert.Truef(t, same, "%s & %s are different files", original, copy)
			}

			Assert.Equal(expectedOut, slices.Collect(iter.Map((strings.Lines(string(stdout))), strings.TrimSpace)))
		})
	}
}
