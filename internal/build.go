package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Cleans up the tmp directory created by Build()
func RemoveBuildDir(bin string) {
	buildDir := filepath.Dir(bin)
	if err := os.RemoveAll(buildDir); err != nil {
		msg := fmt.Sprintf("cannot remove temporary directory used for build output: %v", err)
		panic(msg)
	}
}

// Build current dir with optional tags, to tmp directory, returns path to resulting binary.
//
// Remember to `defer RemoveBuildDir()`
func Build(tags []string) string {
	_, caller, _, _ := runtime.Caller(1)
	buildDir, err := os.MkdirTemp(os.TempDir(), filepath.Base(caller))
	if err != nil {
		panic("Cannot create temporary directory for build output")
	}

	args := []string{"build"}
	if len(tags) > 0 {
		args = append(args, "-tags", strings.Join(tags, ","))
	}
	args = append(args, "-o", buildDir, filepath.Dir(caller))

	if err := exec.Command("go", args...).Run(); err != nil {
		msg := fmt.Sprintf("cannot go %s: %v", args, err)
		panic(msg)
	}

	return filepath.Join(buildDir, "snaggle")
}
