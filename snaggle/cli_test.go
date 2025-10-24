package main

import (
	"fmt"
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

			Assert.ElementsMatch(expectedOut, slices.Collect(iter.Map((strings.Lines(string(stdout))), strings.TrimSpace)))
		})
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
