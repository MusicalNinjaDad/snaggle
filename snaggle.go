package snaggle

import (
	"github.com/u-root/u-root/pkg/ldd"
	"os"
	"path/filepath"
	"slices"
)

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}

func SymlinkTree(path string) (map[string]string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	tree := make(map[string]string)
	leaf, err := os.Readlink(path)
	tree[path] = leaf
	return tree, err
}
