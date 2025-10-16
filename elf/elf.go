package elf

import "path/filepath"

// A parsed Elf binary
type Elf struct {
	Path string // Absolute, fully resolved path to the file
}

func New(path string) (Elf, error) {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return Elf{path}, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return Elf{path}, err
	}
	return Elf{path}, nil
}
