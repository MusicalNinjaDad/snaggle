// This is not designed work on non-linux systems - don't try it unless you want to have fun with unexpected
// and unhandled os error types.

package snaggle

import (
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Non-recursively identifies dynamic libraries for the ELF file at `elfpath` and returns a slice with libraries
// as defined in the ELF followed by the interpreter.
//
// Note:
//   - The interpreter will always be the last entry and will have a full, absolute path.
//     See https://gist.github.com/x0nu11byt3/bcb35c3de461e5fb66173071a2379779 for much more background
//   - The libraries will, hopefully, not include paths. See [man ld.so(8)](https://man7.org/linux/man-pages/man8/ld.so.8.html)
//     for information on how the interpreter searches for the actual files.
//   - An error, containing the phrase "Did not read full interpreter path." will be provided if we are not confident of the interpreter path.
//   - Other errors from `debug.elf` will be propogated as is and a best-effort set of libraries returned.
func LibPaths(elfpath string) ([]string, error) {
	elffile, err := elf.Open(elfpath)
	if err != nil {
		return nil, err
	}

	libs, err := elffile.ImportedLibraries()
	if err != nil {
		return libs, err
	}
	slices.Sort(libs)

	for _, prog := range elffile.Progs {
		if prog.Type == elf.PT_INTERP {
			size := prog.Filesz - 1 // -1 to strip `\x00` termination
			p := make([]byte, size)
			read, err := prog.Open().Read(p)
			if err != nil {
				return libs, err
			}
			libs = append(libs, string(p))
			if uint64(read) != size {
				msg := fmt.Sprint("Did not read full interpreter path. Expected", size, ", read", read, "bytes")
				err := errors.New(msg)
				return libs, err
			}
		}
	}
	return libs, nil
}

// LinkFile creates a hardlink to `path` under `newRoot`, preserving the full directory
// structure similar to how `cp -r` does.
//
// E.g. `LinkFile(/usr/bin/which, /tmp)` will create a link at `/tmp/usr/bin/which`.
//
// Note: the _absolute_ `path` will be used, even if a relative path is provided.
//
// Errors may be propogated from `filepath.Abs`, `os.MkdirAll` and `os.Link`, sadly a complete
// lack of meaningful documentation on stdlib errors means the author of this code can't give
// any guidance on what they may be.
// PRs are always welcome to improve error handling or documentation.
func LinkFile(path string, newRoot string) (string, error) {
	path, err := filepath.Abs(path)
	target := filepath.Join(newRoot, path)
	if err != nil {
		// propogate err from filepath.Abs, with maybe some sort of target
		return target, err
	}

	err = os.MkdirAll(filepath.Dir(target), os.ModePerm)
	if err != nil {
		return target, err
	}

	// TODO: handle err (e.g. "operation not permitted")
	// TODO: what if bin is a symlink?
	err = os.Link(path, target)
	return target, err
}
