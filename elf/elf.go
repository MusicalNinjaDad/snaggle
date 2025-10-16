package elf

import (
	debug_elf "debug/elf"
	"path/filepath"
)

// A parsed Elf binary
type Elf struct {
	Path  string   // Absolute, fully resolved path to the file
	Class EI_CLASS // See https://man7.org/linux/man-pages/man5/elf.5.html#:~:text=.%20%20(3%3A%20%27F%27)-,EI_CLASS,-The%20fifth%20byte
}

type EI_CLASS byte

const (
	ELFNONE = debug_elf.ELFCLASSNONE
	ELF32   = debug_elf.ELFCLASS32
	ELF64   = debug_elf.ELFCLASS64
)

func New(path string) (Elf, error) {
	elf := Elf{path, EI_CLASS(ELFNONE)}
	path, err := filepath.EvalSymlinks(path)
	elf.Path = path
	if err != nil {
		return elf, err
	}
	path, err = filepath.Abs(path)
	elf.Path = path
	if err != nil {
		return elf, err
	}

	elffile, err := debug_elf.Open(path)
	if err != nil {
		return elf, err
	}

	elf.Class = EI_CLASS(elffile.Class)

	return elf, nil
}
