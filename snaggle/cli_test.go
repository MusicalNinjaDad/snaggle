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
	snaggleBin   string
	snagglePanic string
)

func TestMain(m *testing.M) {
	_, thisfile, _, _ := runtime.Caller(0)
	buildTmp, err := os.MkdirTemp(os.TempDir(), filepath.Base(thisfile))
	if err != nil {
		panic("Cannot create temporary directory for build output")
	}
	defer func() {
		if err := os.RemoveAll(buildTmp); err != nil {
			msg := fmt.Sprintf("cannot remove temporary directory used for build output: %v", err)
			panic(msg)
		}
	}()

	snaggleBin = filepath.Join(buildTmp, "snaggle")
	build := exec.Command("go", "build", "-o", snaggleBin, filepath.Dir(thisfile))
	if err := build.Run(); err != nil {
		msg := fmt.Sprintf("cannot %s: %v", build.Args, err)
		panic(msg)
	}

	snagglePanic = filepath.Join(buildTmp, "snaggle-panic")
	build = exec.Command("go", "build", "-tags", "testpanic", "-o", snagglePanic, filepath.Dir(thisfile))
	if err := build.Run(); err != nil {
		msg := fmt.Sprintf("cannot %s: %v", build.Args, err)
		panic(msg)
	}

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

			Assert.Equal(expectedOut, slices.Collect(iter.Map((strings.Lines(string(stdout))), strings.TrimSpace)))
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
	_, thisfile, _, _ := runtime.Caller(0)
	buildTmp, err := os.MkdirTemp(os.TempDir(), filepath.Base(thisfile))
	if err != nil {
		panic("Cannot create temporary directory for build output")
	}
	build := exec.Command("go", "build", "-tags", "testpanic", "-o", buildTmp, filepath.Dir(thisfile))
	if err := build.Run(); err != nil {
		msg := fmt.Sprintf("cannot %s: %v", build.Args, err)
		panic(msg)
	}
	snaggleBin = filepath.Join(buildTmp, "snaggle")

	Assert := assert.New(t)
	snaggle := exec.Command(snaggleBin, "src", "dst")
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
