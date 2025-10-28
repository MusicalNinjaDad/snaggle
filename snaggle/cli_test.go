package main

import (
	"fmt"
	"maps"
	"os"
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

var (
	snaggleBin        string
	_, thisfile, _, _ = runtime.Caller(0)
)

func removeBuildDir(bin string) {
	buildDir := filepath.Dir(bin)
	if err := os.RemoveAll(buildDir); err != nil {
		msg := fmt.Sprintf("cannot remove temporary directory used for build output: %v", err)
		panic(msg)
	}
}

func build(tags []string) string {
	buildDir, err := os.MkdirTemp(os.TempDir(), filepath.Base(thisfile))
	if err != nil {
		panic("Cannot create temporary directory for build output")
	}

	args := []string{"build"}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	args = append(args, "-o", buildDir, filepath.Dir(thisfile))

	if err := exec.Command("go", args...).Run(); err != nil {
		msg := fmt.Sprintf("cannot go %s: %v", args, err)
		panic(msg)
	}

	return filepath.Join(buildDir, "snaggle")

}

func TestMain(m *testing.M) {
	snaggleBin = build(nil)
	defer removeBuildDir(snaggleBin)
	m.Run()
}

func TestCommonBinaries(t *testing.T) {
	tests := CommonBinaries(t)

	for _, inplace := range []bool{false, true} {
		for _, tc := range tests {
			testname := tc.Description
			if inplace {
				testname += "_inplace"
			}
			t.Run(testname, func(t *testing.T) {
				Assert := assert.New(t)
				tmp := WorkspaceTempDir(t)
				snaggle := exec.Command(snaggleBin)

				if inplace {
					snaggle.Args = append(snaggle.Args, "--in-place")
				}
				snaggle.Args = append(snaggle.Args, tc.Elf.Path, tmp)

				expectedOut, expectedFiles := ExpectedOutput(tc, tmp, inplace)
				stdout, err := snaggle.Output()

				if !Assert.NoError(err) {
					var exiterr *exec.ExitError
					Assert.ErrorAs(err, &exiterr)
					t.Logf("Stderr: %s", exiterr.Stderr)
				}

				for original, copy := range expectedFiles {
					if original == tc.Elf.Path {
						AssertLinkedFile(t, original, copy)
					} else {
						AssertSameFile(t, original, copy)
					}
				}

				AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
				Assert.ElementsMatch(expectedOut, StripLines(string(stdout)))
			})
		}
	}
}

func TestInvalidNumberArgs(t *testing.T) {
	Assert := assert.New(t)
	snaggle := exec.Command(snaggleBin, "src")
	out, err := snaggle.Output()
	t.Logf("Stdout: %s", out)
	var exitcode *exec.ExitError
	Assert.ErrorAs(err, &exitcode)
	t.Logf("Stderr: %s", exitcode.Stderr)
	Assert.Equal(2, exitcode.ExitCode())
}

func TestPanic(t *testing.T) {
	panicBin := build([]string{"testpanic"})
	defer removeBuildDir(panicBin)

	Assert := assert.New(t)
	snaggle := exec.Command(panicBin, "src", "dst")
	out, err := snaggle.Output()
	t.Logf("Stdout: %s", out)
	var exitcode *exec.ExitError
	Assert.ErrorAs(err, &exitcode)
	stderr := slices.Collect(iter.Map((strings.Lines(string(exitcode.Stderr))), strings.TrimSpace))
	t.Logf("Stderr: %s", stderr)
	Assert.Equal(3, exitcode.ExitCode())
	Assert.Contains(stderr, "Sorry someone panicked!")
	Assert.Contains(stderr, "This is what we know ...")
	Assert.Contains(stderr, "you got a special testing build that always panics. (Tip: don't build with `-tags testpanic`)")
}

func TestDirectory(t *testing.T) {
	for _, recursive := range []bool{false, true} {
		var testname string
		if recursive {
			testname = "recursive"
		} else {
			testname = "flat"
		}

		t.Run(testname, func(t *testing.T) {
			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)
			dir := TestdataPath(".")

			snaggle := exec.Command(snaggleBin)
			if recursive {
				snaggle.Args = append(snaggle.Args, "--recursive")
			}
			snaggle.Args = append(snaggle.Args, dir, tmp)

			contents := CommonBinaries(t)
			if recursive {
				contents["subdir"] = Hello_dynamic
			}

			var expectedOut []string
			var expectedFiles = make(map[string]string)
			for _, bin := range contents {
				stdout, files := ExpectedOutput(bin, tmp, inplace)
				expectedOut = append(expectedOut, stdout...)
				maps.Insert(expectedFiles, maps.All(files))
			}

			stdout, err := snaggle.Output()

			if !Assert.NoError(err) {
				var exiterr *exec.ExitError
				Assert.ErrorAs(err, &exiterr)
				t.Logf("Stderr: %s", exiterr.Stderr)
			}

			for original, copy := range expectedFiles {
				AssertSameFile(t, original, copy)
			}

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			Assert.ElementsMatch(expectedOut, StripLines(string(stdout)))
		})
	}
}

func TestInvalidElf(t *testing.T) {
	tc := Ldd

	for _, inplace := range []bool{false, true} {
		var testname string
		if inplace {
			testname = "inplace"
		} else {
			testname = "link"
		}

		t.Run(testname, func(t *testing.T) {
			Assert := assert.New(t)
			tmp := WorkspaceTempDir(t)

			snaggle := exec.Command(snaggleBin)

			if inplace {
				snaggle.Args = append(snaggle.Args, "--in-place")
			}
			snaggle.Args = append(snaggle.Args, tc.Elf.Path, tmp)

			expectedOut := make([]string, 0)
			expectedErr := []string{
				"Error: parsing " + tc.Elf.Path + ":",
				"invalid ELF file: bad magic number '[35 33 47 117]' in record at byte 0x0",
				"",
			}
			expectedFiles := make(map[string]string, 0)

			stdout, err := snaggle.Output()

			if Assert.Error(err) {
				var exiterr *exec.ExitError
				Assert.ErrorAs(err, &exiterr)
				Assert.Equal(strings.Join(expectedErr, "\n"), string(exiterr.Stderr))
				Assert.Equal(1, exiterr.ExitCode())
			}

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			Assert.ElementsMatch(expectedOut, StripLines(string(stdout)))
		})
	}
}
