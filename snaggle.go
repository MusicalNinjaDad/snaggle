package snaggle

import "github.com/u-root/u-root/pkg/ldd"

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	return libs, err
}