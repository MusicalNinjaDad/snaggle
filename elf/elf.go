package elf

import (
	"bytes"
	debug_elf "debug/elf"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"slices"
)

// A parsed Elf binary
type Elf struct {
	// The filename
	Name string

	// Absolute, fully resolved path to the file
	Path string

	// 32 or 64 bit?
	//  - See https://man7.org/linux/man-pages/man5/elf.5.html#:~:text=.%20%20(3%3A%20%27F%27)-,EI_CLASS,-The%20fifth%20byte
	Class EI_CLASS

	// Simplified based on ET_DYN & DynFlag1
	Type Type

	// Absolute path to the interpreter (if executable), "" if not executable.
	//  - See https://gist.github.com/x0nu11byt3/bcb35c3de461e5fb66173071a2379779 for much more background
	Interpreter string

	// Names of all requested libraries
	Dependencies []string
}

type EI_CLASS byte

const (
	ELFNONE = debug_elf.ELFCLASSNONE // 0
	ELF32   = debug_elf.ELFCLASS32   // 1
	ELF64   = debug_elf.ELFCLASS64   // 2
)

// Think before directly comparing to bitmask (2^n) values. See value description for individual hints.
type Type byte

const (
	// Bitmask values
	// --------------

	UNDEF = 0 // Undefined
	EXE   = 1 // Executable: Use Elf.IsExe() to catch _any_ type of executable
	DYN   = 2 // Dynamic: Use Elf.IsDyn() to catch _any_ type of dynamically linked binary

	// Meaningful combination values
	// -----------------------------

	PIE = 3 // EXE + DYN
)

func (e *Elf) IsExe() bool {
	return e.Type&Type(EXE) != 0
}

func (e *Elf) IsDyn() bool {
	return e.Type&Type(DYN) != 0
}

func New(path string) (Elf, error) {
	elf := Elf{Path: path}
	var elffile *debug_elf.File
	var errs []error
	var err error

	appenderr := func(err error, message string) {
		err = fmt.Errorf("%s %s: %w", message, elf.Path, err)
		errs = append(errs, err)
	}

	elf.Name = filepath.Base(path)

	elf.Path, err = resolve(path)
	if err != nil {
		return elf, err
	}

	elffile, err = debug_elf.Open(elf.Path)
	if err != nil {
		return elf, err
	}
	defer func() {
		err := elffile.Close()
		if err != nil {
			appenderr(err, "error closing")
		}
	}()

	elf.Class = EI_CLASS(elffile.Class)

	elf.Interpreter, err = interpreter(elffile)
	if err != nil {
		appenderr(err, "error getting interpreter for")
	}

	elf.Type, err = elftype(elffile)
	if err != nil {
		appenderr(err, "error getting type of")
	}

	if elf.Type == Type(PIE) && elf.Interpreter == "" {
		msg := fmt.Sprint(elf.Path, " is a PIE without interpreter")
		err = errors.New(msg)
		errs = append(errs, err)
	}

	elf.Dependencies, err = elffile.ImportedLibraries()
	if err != nil {
		appenderr(err, "error getting dependecies for")
	}
	slices.Sort(elf.Dependencies)

	return elf, errors.Join(errs...)
}

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

// Identifies the type of Elf (binary vs library) based upon a combination of `DT_FLAGS_1` & the claimed `e_type` in the header.
//
//   - Returns `Type(UNDEF), errors.ErrUnsupported` for types we don't recognise.
func elftype(elffile *debug_elf.File) (Type, error) {
	switch claimedtype := elffile.Type; claimedtype {

	case debug_elf.ET_EXEC:
		return Type(EXE), nil

	case debug_elf.ET_DYN:
		pie, err := hasDT_FLAGS_1(elffile, debug_elf.DF_1_PIE)
		if err != nil {
			return Type(DYN), err
		}
		if pie {
			return Type(PIE), nil
		} else {
			return Type(DYN), nil
		}

	default:
		return Type(UNDEF), fmt.Errorf("unsupported elf type: %w", errors.ErrUnsupported)
	}
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

func hasDT_FLAGS_1(elffile *debug_elf.File, flag debug_elf.DynFlag1) (bool, error) {
	dt_flags_1, err := elffile.DynValue(debug_elf.DynTag(debug_elf.DT_FLAGS_1))
	if err != nil {
		return false, fmt.Errorf("error getting DT_FLAGS_1: %w", err)
	}
	for _, flags := range dt_flags_1 {
		// Bitmask against PIE Flag (0x08000000)
		if flags&uint64(flag) != 0 {
			return true, nil
		}
	}
	return false, nil
}
