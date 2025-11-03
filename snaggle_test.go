package snaggle_test

import (
	"io"
	"io/fs"
	"log"
	"maps"
	"os"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MusicalNinjaDad/snaggle"
	"github.com/MusicalNinjaDad/snaggle/elf"
	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

func Test(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	for t, tc := range TestCases(t) {
		t.Cleanup(func() { stdout.Reset() })
		Assert := Assert(t)

		err := snaggle.Snaggle(tc.Src, tc.Dest, tc.Options...)

		Assert.Testify.NoError(err)

		Assert.DirectoryContents(tc.ExpectedFiles, tc.Dest)
		Assert.LinkedFile(tc.Src, tc.ExpectedFiles[tc.Src])

		Assert.Stdout(tc.ExpectedStdout, StripLines(stdout.String()), tc.Src)
	}
}

func BenchmarkCommonBinaries(b *testing.B) {
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(os.Stdout) })

	for _, tc := range GoodElfs {
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
	tmp := WorkspaceTempDir(t)
	for range 2 {
		err := snaggle.Snaggle(P_which, tmp)
		Assert.NoError(err)
	}
}

func TestInvalidElf(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tc := StrangeElfs["ldd"]

	for _, inplace := range []bool{false, true} {
		var testname string
		if inplace {
			testname = "inplace"
		} else {
			testname = "link"
		}

		t.Run(testname, func(t *testing.T) {
			t.Cleanup(func() { stdout.Reset() })

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			var ErrorType *snaggle.SnaggleError

			expectedOut := make([]string, 0)
			expectedFiles := make(map[string]string, 0)

			var err error
			switch {
			case inplace:
				err = snaggle.Snaggle(tc.Elf.Path, tmp, snaggle.InPlace())
			default:
				err = snaggle.Snaggle(tc.Elf.Path, tmp)
			}

			// In CLI test assert StdErr & exit code instead
			if Assert.ErrorAs(err, &ErrorType) {
				Assert.Equal(tc.Elf.Path, ErrorType.Src)
				Assert.Equal("", ErrorType.Dst)
				Assert.ErrorIs(err, elf.ErrInvalidElf)
				Assert.ErrorContains(err, tc.Elf.Path)
			}

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			Assert.ElementsMatch(expectedOut, StripLines(stdout.String()))
		})
	}
}

func TestRecurseFile(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tc := StrangeElfs["ldd"]

	for _, inplace := range []bool{false, true} {
		var testname string
		if inplace {
			testname = "inplace"
		} else {
			testname = "link"
		}

		t.Run(testname, func(t *testing.T) {
			t.Cleanup(func() { stdout.Reset() })

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			var invocationError *snaggle.InvocationError
			var pathError *fs.PathError
			var snaggleError *snaggle.SnaggleError

			expectedOut := make([]string, 0)
			expectedFiles := make(map[string]string, 0)

			var err error
			switch {
			case inplace:
				err = snaggle.Snaggle(tc.Elf.Path, tmp, snaggle.InPlace(), snaggle.Recursive())
			default:
				err = snaggle.Snaggle(tc.Elf.Path, tmp, snaggle.Recursive())
			}

			// In CLI test assert StdErr & exit code instead
			if Assert.ErrorAs(err, &invocationError) {
				Assert.Equal(tc.Elf.Path, invocationError.Path)
				Assert.Equal(tmp, invocationError.Target)
				if Assert.ErrorAs(err, &pathError) {
					Assert.Equal("--recursive", pathError.Op)
					Assert.Equal(tc.Elf.Path, pathError.Path)
					Assert.ErrorIs(pathError, syscall.ENOTDIR)
				}
			}

			Assert.NotErrorAs(err, &snaggleError)
			Assert.NotErrorIs(err, elf.ErrInvalidElf)

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			Assert.ElementsMatch(expectedOut, StripLines(stdout.String()))
		})
	}
}
