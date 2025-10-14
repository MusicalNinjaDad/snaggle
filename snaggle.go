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

func CopyBinary(bin string, newRoot string) (string, error) {
	bin, err := filepath.Abs(bin)
	target := filepath.Join(newRoot, bin)
	if err != nil {
		return target, err
	}

	os.MkdirAll(filepath.Dir(target), os.ModePerm)

	// TODO: handle err (e.g. "operation not permitted")
	// TODO: what if bin is a symlink?
	err = os.Link(bin, target)
	return target, err
}
