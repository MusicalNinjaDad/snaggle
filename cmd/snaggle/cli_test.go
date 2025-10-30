package main

import (
	"maps"
	"os/exec"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/MusicalNinjaDad/snaggle/internal"
	. "github.com/MusicalNinjaDad/snaggle/internal/testing"
)

var snaggleBin string

func TestMain(m *testing.M) {
	snaggleBin = Build(nil)
	defer RemoveBuildDir(snaggleBin)
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

	expectedErr := "Error: snaggle expects 2 argument(s), 1 received\n"
	expectedErr += rootCmd.UsageString()
	expectedErr += "\n"

	stdout, err := snaggle.Output()

	Assert.Empty(stdout)

	var exitError *exec.ExitError
	if Assert.ErrorAs(err, &exitError) {
		Assert.Equal(2, exitError.ExitCode())
		Assert.Equal(expectedErr, string(exitError.Stderr))
	}

	t.Logf("Stdout:\n%s", stdout)
	t.Logf("Stderr:\n%s", exitError.Stderr)
}

func TestPanic(t *testing.T) {
	panicBin := Build([]string{"testpanic"})
	defer RemoveBuildDir(panicBin)

	Assert := assert.New(t)

	snaggle := exec.Command(panicBin, "src", "dst")

	expectedErr := "Sorry someone panicked!\n"
	expectedErr += "This is what we know ...\n"
	expectedErr += "you got a special testing build that always panics. (Tip: don't build with `-tags testpanic`)\n"

	stdout, err := snaggle.Output()

	Assert.Empty(stdout)

	var exitError *exec.ExitError
	if Assert.ErrorAs(err, &exitError) {
		Assert.Equal(3, exitError.ExitCode())
		Assert.True(strings.HasPrefix(string(exitError.Stderr), expectedErr), "stderr does not start as expected")
		Assert.NotContains(string(exitError.Stderr), rootCmd.UsageString())
	}

	t.Logf("Stdout:\n%s", stdout)
	t.Logf("Stderr:\n%s", exitError.Stderr)
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
				t.Logf("Stderr:\n%s", exiterr.Stderr)
			}

			AssertDirectoryContents(t, slices.Collect(maps.Values(expectedFiles)), tmp)
			Assert.ElementsMatch(expectedOut, StripLines(string(stdout)))
		})
	}
}

func TestRecurseFile(t *testing.T) {
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
			snaggle.Args = append(snaggle.Args, "--recursive", tc.Elf.Path, tmp)

			expectedErr := "Error: --recursive " + tc.Elf.Path + ": not a directory\n"
			expectedErr += rootCmd.UsageString()
			expectedErr += "\n"

			stdout, err := snaggle.Output()

			Assert.Empty(stdout)

			var exitError *exec.ExitError
			if Assert.ErrorAs(err, &exitError) {
				Assert.Equal(2, exitError.ExitCode())
				Assert.Equal(expectedErr, string(exitError.Stderr))
			}

			AssertDirectoryContents(t, nil, tmp)

			t.Logf("Stdout:\n%s", stdout)
			t.Logf("Stderr:\n%s", exitError.Stderr)
		})
	}
}
