package main

import (
	"os/exec"
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

func Test(t *testing.T) {
	for t, tc := range TestLoop(t) {
		Assert := Assert(t)

		snaggle := exec.Command(snaggleBin, tc.Flags...)
		snaggle.Args = append(snaggle.Args, tc.Src, tc.Dest)

		stdout, err := snaggle.Output()

		if !Assert.Testify.NoError(err) {
			var exiterr *exec.ExitError
			Assert.Testify.ErrorAs(err, &exiterr)
			t.Logf("Stderr: %s", exiterr.Stderr)
		}

		Assert.DirectoryContents(tc.ExpectedFiles, tc.Dest)
		Assert.LinkedFile(tc.Src, tc.ExpectedFiles[tc.Src])

		Assert.Stdout(tc.ExpectedStdout, StripLines(string(stdout)), tc.Src)

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
func TestRecurseFile(t *testing.T) {
	tests := []TestDetails{TestData[P_ldd]}
	for t, tc := range TestLoop(t, tests...) {
		Assert := Assert(t)

		expectedErr := "Error: --recursive " + tc.Src + ": not a directory\n"
		expectedErr += rootCmd.UsageString()
		expectedErr += "\n"

		tc.Flags = append(tc.Flags, "--recursive")

		snaggle := exec.Command(snaggleBin, tc.Flags...)
		snaggle.Args = append(snaggle.Args, tc.Src, tc.Dest)

		stdout, err := snaggle.Output()

		Assert.Testify.Empty(stdout)
		Assert.DirectoryContents(nil, tc.Dest)

		t.Logf("Stdout:\n%s", stdout)

		if Assert.Testify.Error(err) {
			var exitError *exec.ExitError
			if Assert.Testify.ErrorAs(err, &exitError) {
				Assert.Testify.Equal(2, exitError.ExitCode())
				Assert.Testify.Equal(expectedErr, string(exitError.Stderr))
			}
			t.Logf("Stderr:\n%s", exitError.Stderr)
		}
	}
}

func TestCopyInplace(t *testing.T) {
	Assert := Assert(t)

	src := TestdataPath(".")
	dest := WorkspaceTempDir(t)

	expectedErr := "Error: cannot copy in-place\n"
	expectedErr += rootCmd.UsageString()
	expectedErr += "\n"

	snaggle := exec.Command(snaggleBin, "--copy", "--in-place")
	snaggle.Args = append(snaggle.Args, src, dest)

	stdout, err := snaggle.Output()

	Assert.Testify.Empty(stdout)
	Assert.DirectoryContents(nil, dest)

	t.Logf("Stdout:\n%s", stdout)

	if Assert.Testify.Error(err) {
		var exitError *exec.ExitError
		if Assert.Testify.ErrorAs(err, &exitError) {
			Assert.Testify.Equal(2, exitError.ExitCode())
			Assert.Testify.Equal(expectedErr, string(exitError.Stderr))
		}
		t.Logf("Stderr:\n%s", exitError.Stderr)
	}
}

func TestNotAnELF(t *testing.T) {
	tests := []TestDetails{TestData[P_empty], TestData[P_ldd]}
	for t, tc := range TestLoop(t, tests...) {
		Assert := Assert(t)

		expectedErr := "Error: parsing " + tc.Src + ":\n"
		expectedErr += "invalid ELF file: bad magic number '[35 33 47 117]' in record at byte 0x0\n"

		snaggle := exec.Command(snaggleBin, tc.Flags...)
		snaggle.Args = append(snaggle.Args, tc.Src, tc.Dest)

		stdout, err := snaggle.Output()

		Assert.Testify.Empty(stdout)
		Assert.DirectoryContents(nil, tc.Dest)

		t.Logf("Stdout:\n%s", stdout)

		if Assert.Testify.Error(err) {
			var exitError *exec.ExitError
			if Assert.Testify.ErrorAs(err, &exitError) {
				Assert.Testify.Equal(1, exitError.ExitCode())
				Assert.Testify.Equal(expectedErr, string(exitError.Stderr))
			}
			t.Logf("Stderr:\n%s", exitError.Stderr)
		}
	}
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
