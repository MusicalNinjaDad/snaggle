package snaggle_test

import (
	"io"
	"log"
	"maps"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle"
	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func TestCommonBinaries(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tests := CommonBinaries(t)

	for _, inplace := range []bool{false, true} {
		for _, tc := range tests {
			testname := tc.Description
			if inplace {
				testname += "_inplace"
			}

			t.Run(testname, func(t *testing.T) {
				t.Cleanup(func() { stdout.Reset() })

				Assert := assert.New(t)
				tmp := WorkspaceTempDir(t)

				expectedOut, expectedFiles := ExpectedOutput(tc, tmp, inplace)
				err := snaggle.Snaggle(tc.Elf.Path, tmp)

				Assert.NoError(err)
				for original, copy := range expectedFiles {
					if original == tc.Elf.Path {
						AssertLinkedFile(t, original, copy)
						Assert.Falsef(inplace, "%s was snagged to %s despite inplace snaggle", original, copy)
					} else {
						AssertSameFile(t, original, copy)
					}
				}

				AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
				Assert.ElementsMatch(expectedOut, StripLines(stdout.String()))
			})
		}
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
				if err := snaggle.Snaggle(tc.Elf.Path, tmp); err != nil {
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
		err := snaggle.Snaggle(tc.Elf.Path, tmp)
		Assert.NoError(err)
	}
}
