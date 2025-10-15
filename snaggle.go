package snaggle

import (
	"os"
	"path/filepath"
	"slices"

	"github.com/u-root/u-root/pkg/ldd"
)

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}

// LinkFile creates a hardlink to `path` under `newRoot`, preserving the full directory
// structure similar to how `cp -r` does.
//
// E.g. `LinkFile(/usr/bin/which, /tmp)` will create a link at `/tmp/usr/bin/which`.
//
// Note: the _absolute_ `path` will be used, even if a relative path is provided.
func LinkFile(path string, newRoot string) (string, error) {
	path, err := filepath.Abs(path)
	target := filepath.Join(newRoot, path)
	if err != nil {
		return target, err
	}

	os.MkdirAll(filepath.Dir(target), os.ModePerm)

	// TODO: handle err (e.g. "operation not permitted")
	// TODO: what if bin is a symlink?
	err = os.Link(path, target)
	return target, err
}
