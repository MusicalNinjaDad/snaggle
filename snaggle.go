package snaggle

import (
	"slices"

	"github.com/u-root/u-root/pkg/ldd"
)

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}