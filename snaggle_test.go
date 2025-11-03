package snaggle_test

import (
	"io"
	"io/fs"
	"log"
	"os"
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
		Assert.NoSymlinks(tc.Dest)

		Assert.Stdout(tc.ExpectedStdout, StripLines(stdout.String()), tc.Src)
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

func TestNotAnELF(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tests := []TestDetails{
		{
			Name: "ldd",
			Path: P_ldd,
			Bin:  Ldd,
		},
	}
	for t, tc := range TestCases(t, tests...) {
		t.Cleanup(func() { stdout.Reset() })
		Assert := Assert(t)

		expectedOut := make([]string, 0)
		expectedFiles := make(map[string]string, 0)

		err := snaggle.Snaggle(tc.Src, tc.Dest, tc.Options...)

		var snaggleError *snaggle.SnaggleError
		if Assert.Testify.ErrorAs(err, &snaggleError) {
			Assert.Testify.Equal(tc.Src, snaggleError.Src)
			Assert.Testify.Equal("", snaggleError.Dst)
		}
		Assert.Testify.ErrorIs(err, elf.ErrInvalidElf)
		Assert.Testify.ErrorContains(err, tc.Src)

		Assert.DirectoryContents(expectedFiles, tc.Dest)
		Assert.Stdout(expectedOut, StripLines(stdout.String()))
	}
}

func TestRecurseFile(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tests := []TestDetails{
		{
			Name: "ldd",
			Path: P_ldd,
			Bin:  Ldd,
		},
	}

	for t, tc := range TestCases(t, tests...) {
		t.Cleanup(func() { stdout.Reset() })
		Assert := Assert(t)

		expectedOut := make([]string, 0)
		expectedFiles := make(map[string]string, 0)

		tc.Options = append(tc.Options, snaggle.Recursive())

		err := snaggle.Snaggle(tc.Src, tc.Dest, tc.Options...)

		var invocationError *snaggle.InvocationError
		if Assert.Testify.ErrorAs(err, &invocationError) {
			Assert.Testify.Equal(tc.Src, invocationError.Path)
			Assert.Testify.Equal(tc.Dest, invocationError.Target)
		}

		var pathError *fs.PathError
		if Assert.Testify.ErrorAs(err, &pathError) {
			Assert.Testify.Equal("--recursive", pathError.Op)
			Assert.Testify.Equal(tc.Src, pathError.Path)
			Assert.Testify.ErrorIs(pathError, syscall.ENOTDIR)
		}

		var snaggleError *snaggle.SnaggleError
		Assert.Testify.NotErrorAs(err, &snaggleError)
		Assert.Testify.NotErrorIs(err, elf.ErrInvalidElf)

		Assert.DirectoryContents(expectedFiles, tc.Dest)
		Assert.Stdout(expectedOut, StripLines(stdout.String()))
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
