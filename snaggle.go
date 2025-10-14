package snaggle

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/u-root/u-root/pkg/ldd"
)

// A symlink or simple file.
//   - Source, Target != "", Err == nil -> Symlink
//   - Source != "", Target == "", Err == nil -> File
//   - Err != nil -> Something else
type Symlink struct {
	Source string
	Target string
	Err    error
}

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}

func NewSymlink(source string) Symlink {
	source, err := filepath.Abs(source)
	if err != nil {
		return Symlink{source, "", err}
	}
	target, err := os.Readlink(source)
	return Symlink{source, target, err}
}
