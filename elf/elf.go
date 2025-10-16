package elf

import (
	"bytes"
	debug_elf "debug/elf"
	"errors"
	"fmt"
	"io"
	"path/filepath"
)

// A parsed Elf binary
type Elf struct {
	Path        string   // Absolute, fully resolved path to the file
	Class       EI_CLASS // 32 or 64 bit? See https://man7.org/linux/man-pages/man5/elf.5.html#:~:text=.%20%20(3%3A%20%27F%27)-,EI_CLASS,-The%20fifth%20byte
	Interpreter string   // Absolute path to the interpreter (if executable), "" if not executable. See https://gist.github.com/x0nu11byt3/bcb35c3de461e5fb66173071a2379779 for much more background
}

type EI_CLASS byte

const (
	ELFNONE = debug_elf.ELFCLASSNONE
	ELF32   = debug_elf.ELFCLASS32
	ELF64   = debug_elf.ELFCLASS64
)

// resolve resolves symlinks and returns an absolute path.
func resolve(path string) (string, error) {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path, err
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return path, err
	}
	return path, nil
}

// Identify the interpreter requested by the ELF, based upon the `PT_INTERP` Program header.
//
// Returns (... ,nil):
//   - `path` if a valid entry was found.
//   - `""` if no such header is present. (E.g. for a library)
//
// Errors:
//
// Errors will include a best-effort value for what we found in the header `(entry, ...)` plus one of the following errors:
//   - "did not read full interpreter path." - if we are not confident to have proprely retrieved the entry
//   - "zero-length entry" - if the entry was present but empty
//   - Anything propogated from io.ReadAll
func interpreter(elffile *debug_elf.File) (string, error) {
	for _, prog := range elffile.Progs {
		if prog.Type == debug_elf.PT_INTERP {
			p := prog.Open()
			interp, err := io.ReadAll(p)
			if err != nil {
				return string(interp), err
			}
			interpreter := string(bytes.TrimRight(interp, "\x00")) // strip `\x00` termination
			if len(interpreter) != int(prog.Filesz-1) {            // have multi-byte chars or unexpected contents
				msg := fmt.Sprint("did not read full interpreter path. Expected", prog.Filesz-1, ", read", len(interpreter), "bytes")
				err := errors.New(msg)
				return string(interp), err
			}
			if len(interpreter) == 0 {
				msg := "zero-length interpreter"
				err := errors.New(msg)
				return string(interp), err
			}
			return interpreter, nil
		}
	}
	return "", nil
}

func New(path string) (Elf, error) {
	elf := Elf{path, EI_CLASS(ELFNONE), ""}
	var elffile *debug_elf.File
	var err error

	elf.Path, err = resolve(path)
	if err != nil {
		return elf, err
	}

	elffile, err = debug_elf.Open(path)
	if err != nil {
		return elf, err
	}

	elf.Class = EI_CLASS(elffile.Class)

	elf.Interpreter, err = interpreter(elffile)
	if err != nil {
		return elf, err
	}

	return elf, nil
}
