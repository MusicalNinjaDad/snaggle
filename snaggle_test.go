package snaggle_test

import (
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/davecgh/go-spew/spew"
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

		if !Assert.Testify.NoError(err) {
			for {
				t.Log(spew.Sdump(err))
				if err = errors.Unwrap(err); err == nil {
					break
				}
			}
		}

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

func TestLinkDifferentFile(t *testing.T) {
	tests := []TestDetails{
		{
			Name: "hello_pie",
			Path: P_hello_pie,
			Bin:  GoodElfs["hello_pie"],
		},
	}

	for t, tc := range TestCases(t, tests...) {
		Assert := Assert(t)

		err := os.MkdirAll(filepath.Join(tc.Dest, "lib64"), 0775)
		Assert.Testify.NoError(err)
		err = os.Link(P_hello_pie, filepath.Join(tc.Dest, P_ld_linux))
		Assert.Testify.NoError(err)
		Assert.LinkedFile(P_hello_pie, filepath.Join(tc.Dest, P_ld_linux))

		err = snaggle.Snaggle(tc.Src, tc.Dest, tc.Options...)

		var linkError *os.LinkError
		if Assert.Testify.ErrorAs(err, &linkError) {
			Assert.Testify.Equal("link", linkError.Op)
			Assert.Testify.Equal(P_ld_linux_resolved, linkError.Old)
			Assert.Testify.Equal(tc.ExpectedFiles[P_ld_linux], linkError.New)
			Assert.Testify.ErrorIs(linkError, syscall.EEXIST)
		}

		var pathError *fs.PathError
		if Assert.Testify.ErrorAs(err, &pathError) {
			Assert.Testify.Equal("link", pathError.Op)
			Assert.Testify.Equal(P_ld_linux, pathError.Path)
			Assert.Testify.Equal(linkError, pathError.Err)
		}

		var snaggleError *snaggle.SnaggleError
		if Assert.Testify.ErrorAs(err, &snaggleError) {
			Assert.Testify.Equal(tc.Src, snaggleError.Src)
			Assert.Testify.Equal(tc.Dest, snaggleError.Dst)
			Assert.Testify.Equal(pathError, snaggleError.Unwrap())
		}
	}

}

func BenchmarkCommonBinaries(b *testing.B) {
	log.SetOutput(io.Discard)
	b.Cleanup(func() { log.SetOutput(os.Stdout) })

	benchtmp := WorkspaceTempDir(b)

	for _, verbose := range []bool{false, true} {
		for _, tc := range GoodElfs {
			desc := tc.Description
			if verbose {
				desc += "_verbose"
			}
			b.Run(desc, func(b *testing.B) {
				i := 0
				for b.Loop() {
					b.StopTimer()
					i++
					tmp, err := os.MkdirTemp(benchtmp, desc)
					if err != nil {
						b.Fatalf("creating %s (%v): %v", tmp, i, err)
					}
					b.Cleanup(func() { _ = os.RemoveAll(tmp) })

					opts := make([]snaggle.Option, 0, 1)
					if verbose {
						opts = append(opts, snaggle.Verbose())
					}

					b.StartTimer()
					if err := snaggle.Snaggle(tc.Elf.Path, tmp, opts...); err != nil {
						b.Fatalf("running %s (%v): %v", desc, i, err)
					}
				}
			})
		}

		desc := "Directory"
		if verbose {
			desc += "_verbose"
		}
		b.Run(desc, func(b *testing.B) {
			i := 0
			src := TestdataPath(".")
			for b.Loop() {
				b.StopTimer()
				i++
				tmp, err := os.MkdirTemp(benchtmp, desc)
				if err != nil {
					b.Fatalf("creating %s (%v): %v", tmp, i, err)
				}
				b.Cleanup(func() { _ = os.RemoveAll(tmp) })

				opts := make([]snaggle.Option, 0, 2)
				opts = append(opts, snaggle.Recursive())
				if verbose {
					opts = append(opts, snaggle.Verbose())
				}

				b.StartTimer()
				if err := snaggle.Snaggle(src, tmp, opts...); err != nil {
					b.Fatalf("running %s (%v): %v", desc, i, err)
				}
			}
		})
	}
}
