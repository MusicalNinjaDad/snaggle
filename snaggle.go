package snaggle

import (
	"github.com/u-root/u-root/pkg/ldd"
	"slices"
)

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}
