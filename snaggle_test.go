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

				var err error
				switch {
				case inplace:
					err = snaggle.Snaggle(tc.Elf.Path, tmp, snaggle.InPlace())
				default:
					err = snaggle.Snaggle(tc.Elf.Path, tmp)
				}

				Assert.NoError(err)

				for original, copy := range expectedFiles {
					if original == tc.Elf.Path {
						AssertLinkedFile(t, original, copy)
					} else {
						AssertSameFile(t, original, copy)
					}
				}

				AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
				AssertStdout(t, expectedOut, StripLines(stdout.String()))
				for _, line := range StripLines(stdout.String()) {
					if strings.Contains(line, tc.Elf.Path) {
						Assert.Conditionf(func() (success bool) {
							return strings.HasPrefix(line, "link ")
						}, "%s should have been linked: %s", tc.Elf.Path, line)
					}
				}
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

func TestDirectory(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	for _, recursive := range []bool{false, true} {
		var testname string
		if recursive {
			testname = "recursive"
		} else {
			testname = "flat"
		}

		t.Run(testname, func(t *testing.T) {
			t.Cleanup(func() { stdout.Reset() })

			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			contents := CommonBinaries(t)
			if recursive {
				contents["subdir"] = Hello_dynamic
			}
			dir := TestdataPath(".")
			inplace := false

			var expectedOut []string
			var expectedFiles = make(map[string]string)

			for _, bin := range contents {
				stdout, files := ExpectedOutput(bin, tmp, inplace)
				expectedOut = append(expectedOut, stdout...)
				maps.Insert(expectedFiles, maps.All(files))
			}

			var err error
			if recursive {
				err = snaggle.Snaggle(dir, tmp, snaggle.Recursive())
			} else {
				err = snaggle.Snaggle(dir, tmp)
			}

			Assert.NoError(err)

			for original, copy := range expectedFiles {
				AssertSameFile(t, original, copy)
			}

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			AssertStdout(t, expectedOut, StripLines(stdout.String()))
		})
	}
}

func TestInvalidElf(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	tc := Ldd

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

	tc := Ldd

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

func Test(t *testing.T) {
	var stdout strings.Builder
	log.SetOutput(&stdout)
	t.Cleanup(func() { log.SetOutput(os.Stdout) })

	for Assert, tc := range TestCases(t) {
		t.Cleanup(func() { stdout.Reset() })

		if !Assert.DirExists(tc.Dest) {
			t.Fatal("Temporary destination does not exist")
		}

		err := snaggle.Snaggle(tc.Src, tc.Dest)

		Assert.NoError(err)

		AssertDirectoryContents(t, slices.Collect(maps.Values(tc.ExpectedFiles)), tc.Dest)
		for original, copy := range tc.ExpectedFiles {
			if original == tc.Src {
				AssertLinkedFile(t, original, copy)
			} else {
				AssertSameFile(t, original, copy)
			}
		}

		AssertStdout(t, tc.ExpectedStdout, StripLines(stdout.String()))
		for _, line := range StripLines(stdout.String()) {
			if strings.Contains(line, tc.Src) {
				Assert.Conditionf(func() (success bool) {
					return strings.HasPrefix(line, "link ")
				}, "%s should have been linked: %s", tc.Src, line)
			}
		}
	}
}
