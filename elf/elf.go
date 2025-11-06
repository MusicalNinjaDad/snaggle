// Parses an ELF binary providing additional details vs. [debug/elf]
//
// # Errors:
//
// All errors returned will be of the type [ErrElf].
//
// # Usage:
//
// Construct a new [Elf] struct with [elf.New]
//
//	bin, err := elf.New(path)
//
//	bin is an Elf with the following structure:
//	{
//		Name: base filename
//		Path: absolute path
//		Class: 32-bit or 64-bit?
//		Type: EXE, BIN, PIE, ...
//		Interpreter: path to requested interpeter
//		Dependencies: slice of paths to dependencies as identified and found by the interpreter
//	}
//
// # Note:
//
// Only accepts static binaries or dynamic binaries which use ld-linux*.so as the interpreter
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

	"github.com/MusicalNinjaDad/snaggle/internal"
)

// All errors returned will be of the type ErrElf.
//   - ErrElf can store multiple errors for a single Elf struct. Use .Join() to add an error.
//   - Will not == nil, even if empty. Use .IsEmpty() to check and then manually return (something, nil)
//
// To extract the path use [errors.As] followed by .Path()
//
//	var errelf *ErrElf
//	if errors.As(err, &errelf) {
//	     errelf.Path()
//	}
//
// .
type ErrElf struct {
	path string
	errs []error
}

// # Primary Errors
var (
	// Error returned when the provided file is not a valid Elf.
	ErrInvalidElf = errors.New("invalid ELF file")
	// Error wrapping a failure when calling `ld-linux*.so` (like `ldd`) to identify dependencies
	ErrLdd = errors.New("ldd failed to execute")
)

// # Specific errors which wrap [ErrInvalidElf]
var (
	// Error returned if dynamic ELF has a bad entry for the interpreter
	ErrBadInterpreter = fmt.Errorf("%w: bad interpreter", ErrInvalidElf)
	// Error returned if the ELF is not a type we support (currently only ET_EXEC & ET_DYN)
	ErrUnsupportedElfType = fmt.Errorf("%w: %w (unsupported ELF Type)", ErrInvalidElf, errors.ErrUnsupported)
	// Error returned if the interpreter is not `ld-linux*.so`
	ErrUnsupportedInterpreter = fmt.Errorf("%w: %w (unsupported interpreter)", ErrInvalidElf, errors.ErrUnsupported)
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
	// All requested libraries
	Dependencies []string
}

// 32 or 64 bit?
//   - See https://man7.org/linux/man-pages/man5/elf.5.html#:~:text=.%20%20(3%3A%20%27F%27)-,EI_CLASS,-The%20fifth%20byte
type EI_CLASS byte

// # Values for [EI_CLASS]
const (
	ELFNONE = debug_elf.ELFCLASSNONE // 0
	ELF32   = debug_elf.ELFCLASS32   // 1
	ELF64   = debug_elf.ELFCLASS64   // 2
)

// Binary type
//
// Think carefully before directly comparing to bitmask (2^n) values. See value descriptions for individual hints.
type Type byte

// # Bitmask values for [Type]
//
// Think carefully before directly comparing to bitmask (2^n) values. See value descriptions for individual hints.
const (
	UNDEF = 0 // Undefined
	EXE   = 1 // Executable: Use Elf.IsExe() to catch _any_ type of executable
	DYN   = 2 // Dynamic: Use Elf.IsDyn() to catch _any_ type of dynamically linked binary
)

// # Meaningful combination values for [Type]
const (
	PIE = 3 // EXE + DYN
)

// Is this ELF **primarily** an executable.
//
//   - This will return `false` in cases such as `/lib64/ld-linux-x86-64.so.2` which has an entry point
//     and _may_ be `exec()`'ed but is primarily designed as a library and therefore not `ET_EXEC` or `PIE`
func (e *Elf) IsExe() bool {
	return e.Type&Type(EXE) != 0
}

// Is this ELF **primarily** a library?
//
//   - If it's not **primarily** an executable, then it's a library (Slightly simplified but good enough for our case)
func (e *Elf) IsLib() bool {
	return e.Type&Type(EXE) == 0
}

// Is this ELF dynamically linked?
func (e *Elf) IsDyn() bool {
	return e.Type&Type(DYN) != 0
}

// Construct a new [Elf] for the file located at path, any error will be an [ErrElf]
//
//   - Returns a best-effort result on error
//   - Returns early if errors are encountered in resolving the Path or in initial parsing by [debug_elf.Open],
//     in this case Name & Path will be filled, although Path may not be fully resolved
//   - If errors are encountered in parsing these will be collected in the returned [ErrElf] and the result will
//     contain as much valid information as possible
func New(path string) (Elf, error) {
	elf := Elf{Path: path}
	reterr := &ErrElf{path: path} // error(s) returned from this function
	var err error                 // individual error returned by any functions called
	var elffile *debug_elf.File   // the opened File

	elf.Name = filepath.Base(path)

	elf.Path, err = resolve(path)
	if err != nil {
		if elf.Path == "" { // resolve may return "" on error
			elf.Path = path // so we reset the path if that's happened
		}
		reterr.Join(err)
		return elf, reterr
	}

	elffile, err = debug_elf.Open(elf.Path)
	if err != nil {
		var formaterr *debug_elf.FormatError
		if errors.As(err, &formaterr) {
			err = fmt.Errorf("%w: %w", ErrInvalidElf, err)
		}
		reterr.Join(err)
		return elf, reterr
	}
	defer func() {
		reterr.Join(
			elffile.Close(),
		)
	}()

	elf.Class = EI_CLASS(elffile.Class)

	elf.Interpreter, err = interpreter(elffile)
	if err != nil {
		reterr.Join(err)
	}

	elf.Type, err = elftype(elffile)
	if err != nil {
		reterr.Join(err)
	}
	if elf.Type == Type(PIE) && elf.Interpreter == "" {
		err = fmt.Errorf("%w (PIE without interpreter)", ErrBadInterpreter)
		reterr.Join(err)
	}

	if elf.IsDyn() {
		elf.Dependencies, err = ldd(elf.Path, elf.Interpreter)
		if err != nil {
			reterr.Join(err)
		}
	}

	if reterr.IsError() {
		return elf, reterr
	}
	return elf, nil
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
//   - Returns `Type(UNDEF), ErrUnsupportedElfType` for types we don't recognise.
func elftype(elffile *debug_elf.File) (Type, error) {
	switch claimedtype := elffile.Type; claimedtype {

	case debug_elf.ET_EXEC:
		return Type(EXE), nil

	case debug_elf.ET_DYN:
		pie, err := hasDT_FLAGS_1_Flag(elffile, debug_elf.DF_1_PIE)
		if err != nil {
			return Type(DYN), err
		}
		if pie {
			return Type(PIE), nil
		} else {
			return Type(DYN), nil
		}

	default:
		return Type(UNDEF), fmt.Errorf("%w: %s", ErrUnsupportedElfType, claimedtype)
	}
}

// Identify the interpreter requested by the ELF, based upon the `PT_INTERP` Program header.
//
// Returns (... ,nil):
//   - `path` if a valid entry was found.
//   - `""` if no such header is present. (E.g. for a library)
//
// Errors will include a best-effort value for what we found in the header `(entry, ...)` plus one of the following errors:
//   - [ErrBadInterpreter] if the entry is invalid or broken
//   - Anything propogated from [io.ReadAll] with the added prefix "IO error reading interpreter:"
func interpreter(elffile *debug_elf.File) (string, error) {
	for _, prog := range elffile.Progs {
		if prog.Type == debug_elf.PT_INTERP {
			p := prog.Open()
			interp, err := io.ReadAll(p)
			if err != nil {
				err := fmt.Errorf("IO error reading interpreter: %w", err)
				return string(interp), err
			}
			interpreter := string(bytes.TrimRight(interp, "\x00")) // strip `\x00` termination
			if len(interpreter) != int(prog.Filesz-1) {            // have multi-byte chars or unexpected contents
				err := fmt.Errorf("%w: expected %v bytes, read %v bytes (%s)", ErrBadInterpreter, prog.Filesz-1, len(interpreter), interpreter)
				return string(interp), err
			}
			if len(interpreter) == 0 {
				err := fmt.Errorf("%w: zero-length interpreter", ErrBadInterpreter)
				return string(interp), err
			}
			return interpreter, nil
		}
	}
	return "", nil
}

// Does the same as `ldd` under the hood - calls the interpreter with `LD_TRACE_LOADED_OBJECTS=1`;
// then parses the output to return ONLY dependencies which the interpreter had to find.
//
// Note:
//   - Will not return any dependencies which contain a `/`
//   - See: https://man7.org/linux/man-pages/man8/ld.so.8.html for full details of
//     how the search is performed.
//   - In case of error a ErrElfLdd is returned, along with the underlying error(s)
//   - WARNING: does no sanity chacking on the input path - make sure what you are passing refers
//     to a valid dynamically linked ELF, which `ld-linux.so*` can parse. E.g.: passing a statically
//     linked ELF will lead to a segfault (which gets caught and returned as an error).
//   - WARNING: Behaviour is *undefined* for interpreters except `ld-linux.so*`
func ldd(path string, interpreter string) ([]string, error) {
	if interpreter == "" {
		interpreter = internal.P_ld_linux
	} else if !internal.Ld_linux_64_RE.MatchString(interpreter) {
		return nil, fmt.Errorf("%w '%s'", ErrUnsupportedInterpreter, interpreter)
	}

	ldso := exec.Command(interpreter, path)
	ldso.Env = append(ldso.Env, "LD_TRACE_LOADED_OBJECTS=1")
	stdout, err := ldso.Output()
	if err != nil {
		return nil, fmt.Errorf("%w %s %s: %w", ErrLdd, interpreter, path, err)
	}

	dependencies := make([]string, 0, strings.Count(string(stdout), "=>"))
	lines := strings.Lines(string(stdout))
	for line := range lines {
		if strings.Contains(line, "=>") {
			dependencies = append(dependencies, strings.Fields(line)[2])
		}
	}

	slices.SortFunc(dependencies, libpathcmp)
	return dependencies, nil
}

// Does the file have the given DT_FLAGS_1 flag set?
func hasDT_FLAGS_1_Flag(elffile *debug_elf.File, flag debug_elf.DynFlag1) (bool, error) {
	dt_flags_1, err := elffile.DynValue(debug_elf.DynTag(debug_elf.DT_FLAGS_1))
	if err != nil {
		return false, fmt.Errorf("%w: invalid DT_FLAGS_1: %w", ErrInvalidElf, err)
	}
	for _, flags := range dt_flags_1 {
		// Bitmask against PIE Flag (0x08000000)
		if flags&uint64(flag) != 0 {
			return true, nil
		}
	}
	return false, nil
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

		switch field.Name {
		case "Path":
			if libpathcmp(selfVal.(string), otherVal.(string)) != 0 {
				diffs = append(diffs, fmt.Sprintf("%s differs for %s: %v != %v", field.Name, self.FieldByName("Name"), selfVal, otherVal))
			}
		case "Dependencies":
			selfDeps := self.FieldByIndex(field.Index).Interface().([]string)
			otherDeps := other.FieldByIndex(field.Index).Interface().([]string)
			if len(selfDeps) != len(otherDeps) {
				diffs = append(diffs, fmt.Sprintf("%s has %v dependencies in left, %v dependencies in right", self.FieldByName("Name"), len(selfDeps), len(otherDeps)))
			} else {
				for idx, selfDep := range selfDeps {
					otherDep := otherDeps[idx]
					if libpathcmp(selfDep, otherDep) != 0 {
						diffs = append(diffs, fmt.Sprintf("dependency %v differs for %s: %s != %s", idx, self.FieldByName("Name"), selfDep, otherDep))
					}
				}
			}
		default:
			if !reflect.DeepEqual(selfVal, otherVal) {
				diffs = append(diffs, fmt.Sprintf("%s differs for %s: %v != %v", field.Name, self.FieldByName("Name"), selfVal, otherVal))
			}
		}
	}
	return diffs
}

// If both paths are absolute: compares only the filename, otherwise compares the entire path.
func libpathcmp(path1 string, path2 string) int {
	if filepath.IsAbs(path1) && filepath.IsAbs(path2) {
		return strings.Compare(filepath.Base(path1), filepath.Base(path2))
	}
	return strings.Compare(path1, path2)
}

func (e *ErrElf) Error() string {
	if e.IsError() {
		return "parsing " + e.path + ":\n" + errors.Join(e.errs...).Error()
	} else {
		return ""
	}

}

// Does this [ErrElf] contain any errors?
func (e *ErrElf) IsError() bool {
	return len(e.errs) > 0
}

func (e *ErrElf) Unwrap() []error {
	return e.errs
}

// Add an error to the [ErrElf]
func (e *ErrElf) Join(err error) {
	if err != nil {
		e.errs = append(e.errs, err)
	}
}

// The path of the [Elf] which this [ErrElf] describes
func (e *ErrElf) Path() string {
	return e.path
}
