package elf

import (
	"bytes"
	debug_elf "debug/elf"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
)

// All errors returned will satisfy `errors.Is(err, ErrElf)`
var ErrElf = errors.New("error from snaggle/elf")

// Error returned when calling `ld.so` (like `ldd`) to identify dependencies
var ErrElfLdd error

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

	// All requested libraries
	Dependencies []Elf
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

// Whether this is **primarily** an executable.
//
//   - This will return `false` in cases such as `/lib64/ld-linux-x86-64.so.2` which has an entry point
//     and _may_ be `exec()`'ed but is not `ET_EXEC` or `PIE`
func (e *Elf) IsExe() bool {
	return e.Type&Type(EXE) != 0
}

// If it's not **primarily** an executable, then it's a library.
// (Slightly simplified but good enough for our case)
func (e *Elf) IsLib() bool {
	return e.Type&Type(EXE) == 0
}

func (e *Elf) IsDyn() bool {
	return e.Type&Type(DYN) != 0
}

// Deeply check for diffs between two `Elf`s. Ignores differences in the Path, as long as:
//  1. Both `Path`s are absolute
//  2. Both `Path`s end in the same filename
func (e Elf) Diff(o Elf) []string {
	var diffs []string
	elf := reflect.TypeOf(e)
	self := reflect.ValueOf(e)
	other := reflect.ValueOf(o)

	for _, field := range reflect.VisibleFields(elf) {
		selfVal := self.FieldByIndex(field.Index).Interface()
		otherVal := other.FieldByIndex(field.Index).Interface()

		if field.Name == "Path" {
			if filepath.IsAbs(selfVal.(string)) && filepath.IsAbs(otherVal.(string)) {
				selfVal = filepath.Base(selfVal.(string))
				otherVal = filepath.Base(otherVal.(string))
			}
		}

		if field.Name == "Dependencies" {
			selfDeps := self.FieldByIndex(field.Index).Interface().([]Elf)
			otherDeps := other.FieldByIndex(field.Index).Interface().([]Elf)
			if len(selfDeps) != len(otherDeps) {
				diffs = append(diffs, fmt.Sprintf("%s has %v dependencies in left, %v dependencies in right", self.FieldByName("Name"), len(selfDeps), len(otherDeps)))
			} else {
				for idx, dep := range selfDeps {
					diffs = append(diffs, dep.Diff(otherDeps[idx])...)
				}
			}
		} else if !reflect.DeepEqual(selfVal, otherVal) {
			diffs = append(diffs, fmt.Sprintf("%s differs for %s: %v != %v", field.Name, self.FieldByName("Name"), selfVal, otherVal))
		}
	}
	return diffs
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
		err = fmt.Errorf("%s is a PIE without interpreter", elf.Path)
		errs = append(errs, err)
	}

	Dependencies, err := elffile.ImportedLibraries()
	if err != nil {
		appenderr(err, "error getting dependecies for")
	}

	ldso := exec.Command("/lib64/ld-linux-x86-64.so.2", elf.Path)
	ldso.Env = append(ldso.Env, "LD_TRACE_LOADED_OBJECTS=1")
	stdout, err := ldso.Output()
	if err != nil {
		appenderr(err, "error calling ldso on")
	}
	for line := range strings.Lines(string(stdout)) {
		if strings.Contains(line, "=>") {
			println(line)
		}
	}

	slices.Sort(Dependencies)
	for _, dep := range Dependencies {
		depPath := filepath.Join("/lib64", dep)
		depElf, err := New(depPath)
		if err != nil {
			appenderr(err, "error generating dependency for")
		}
		elf.Dependencies = append(elf.Dependencies, depElf)
	}

	return elf, errors.Join(errs...)
}

// resolve resolves symlinks and returns an absolute path.
func resolve(path string) (string, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return path, err
	}
	path, err = filepath.EvalSymlinks(path)
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
				err := fmt.Errorf("did not read full interpreter path: expected %v bytes, read %v bytes", prog.Filesz-1, len(interpreter))
				return string(interp), err
			}
			if len(interpreter) == 0 {
				err := errors.New("zero-length interpreter")
				return string(interp), err
			}
			return interpreter, nil
		}
	}
	return "", nil
}

// Does the same as `ldd` under the hood - calls `ld-linux.so*` with `LD_TRACE_LOADED_OBJECTS=1`;
// then parses the output to return ONLY dependencies which `ld-linux.so*` had to find.
//
// Note:
//   - Will not return any dependencies which contain a `/`
//   - See: https://man7.org/linux/man-pages/man8/ld.so.8.html for full details of
//     how the search is performed.
//   - In case of error a ErrElfLdd is returned, along with the underlying error(s)
//   - WARNING: does no sanity chacking on the input path - make sure what you are passing refers
//     to a valid dynamically linked ELF, which `ld-linux.so*` can parse. E.g.: passing a statically
//     linked ELF will lead to a segfault (which gets caught and returned as an error).
func ldd(path string) ([]string, error) {
	ErrElfLdd = errors.Join(ErrElf, fmt.Errorf("error calling ldso on %s", path))
	ldso := exec.Command("/lib64/ld-linux-x86-64.so.2", path)
	ldso.Env = append(ldso.Env, "LD_TRACE_LOADED_OBJECTS=1")
	stdout, err := ldso.Output()
	if err != nil {
		return nil, errors.Join(ErrElfLdd, err)
	}
	dependencies := make([]string, 0, strings.Count(string(stdout), "=>"))
	lines := strings.Lines(string(stdout))
	for line := range lines {
		if strings.Contains(line, "=>") {
			dependencies = append(dependencies, strings.Fields(line)[2])
		}
	}
	slices.SortFunc(dependencies, sortByFilename)
	return dependencies, err
}

func sortByFilename(a string, b string) int {
	return strings.Compare(filepath.Base(a), filepath.Base(b))
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

// If both paths are absolute: compares only the filename, otherwise compares the entire path.
func libpathcmp(path1 string, path2 string) int {
	if filepath.IsAbs(path1) && filepath.IsAbs(path2) {
		return strings.Compare(filepath.Base(path1), filepath.Base(path2))
	}
	return strings.Compare(path1, path2)
}
