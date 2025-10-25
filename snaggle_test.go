package snaggle_test

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/ameghdadian/x/iter"
	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle"
	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			var stdout strings.Builder
			log.SetOutput(&stdout)
			t.Cleanup(func() { log.SetOutput(os.Stdout) })

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			binPath := filepath.Join(tmp, "bin", filepath.Base(tc.ExpectedElf.Name))
			expectedOut := make([]string, 0, 1+len(tc.ExpectedElf.Dependencies))
			expectedOut = append(expectedOut, tc.ExpectedElf.Path+" -> "+binPath)
			// TODO: ugly - should be in the tc - needs a tidy
			if tc.ExpectedElf.IsDyn() && !tc.ExpectedElf.IsLib() {
				expectedOut = append(expectedOut, tc.ExpectedElf.Interpreter+" -> "+filepath.Join(tmp, P_ld_linux))
			}
			var libCopies []string
			for _, lib := range tc.ExpectedElf.Dependencies {
				copy := filepath.Join(tmp, "lib64", filepath.Base(lib))
				libCopies = append(libCopies, copy)
				expectedOut = append(expectedOut, lib+" -> "+copy)
			}

			err := snaggle.Snaggle(tc.ExpectedElf.Path, tmp)
			Assert.NoError(err)
			AssertSameInode(t, tc.ExpectedElf.Path, binPath)
			for idx, copy := range libCopies {
				original := tc.ExpectedElf.Dependencies[idx]
				same := SameFile(original, copy)
				assert.Truef(t, same, "%s & %s are different files", original, copy)
			}

			Assert.ElementsMatch(expectedOut, slices.Collect(iter.Map((strings.Lines(stdout.String())), strings.TrimSpace)))
		})
	}
}

func BenchmarkCommonBinaries(b *testing.B) {
	tests := CommonBinaries(b)
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(os.Stdout) })

	for _, tc := range tests {
		b.Run(tc.Description, func(b *testing.B) {
			basetmp := WorkspaceTempDir(b)
			i := 0
			for b.Loop() {
				i++
				tmp, err := os.MkdirTemp(basetmp, tc.Description)
				if err != nil {
					b.Fatalf("creating %s (%v): %v", tmp, i, err)
				}
				if err := snaggle.Snaggle(tc.ExpectedElf.Path, tmp); err != nil {
					b.Fatalf("running %s (%v): %v", tc.Description, i, err)
				}
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	Assert := assert.New(t)
	tc := CommonBinaries(t)["PIE_1"]
	tmp := WorkspaceTempDir(t)
	for range 2 {
		err := snaggle.Snaggle(tc.ExpectedElf.Path, tmp)
		Assert.NoError(err)
	}
}
