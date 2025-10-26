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
	Assert := assert.New(t)
	tests := CommonBinaries(t)

	for _, tc := range tests {
		t.Run(tc.Description, func(t *testing.T) {
			tmp := WorkspaceTempDir(t)
			snaggle := exec.Command(snaggleBin, tc.Elf.Path, tmp)

			expectedOut, expectedFiles := ExpectedOutput(tc, tmp)

			stdout, err := snaggle.Output()
			Assert.NoError(err)
			for original, copy := range expectedFiles {
				if original == tc.Elf.Path {
					AssertSameFile(t, original, copy, true)
				} else {
					AssertSameFile(t, original, copy, false)
				}
			}
			Assert.ElementsMatch(expectedOut, StripLines(string(stdout)))
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
