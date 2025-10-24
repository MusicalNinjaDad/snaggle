// Snag a copy of a ELF binary and all its dependencies to another/path/bin & another/path//lib64.
//
// This is the main implementation of the command-line application `snaggle`, for use as a library in
// other code and scripts.
//
// Snaggle is designed to help create minimal runtime containers from pre-existing installations.
// It may work for other use cases and I'd be interested to hear about them at:
// https://github.com/MusicalNinjaDad/snaggle
//
// WARNING: This is not designed work on non-linux systems - don't try it unless you want to have fun
// with unexpected and unhandled os error types.
package snaggle

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/MusicalNinjaDad/snaggle/internal"
)

func init() {
	log.SetFlags(0)
}

// create a hardlink in targetDir which references sourcePath,
// falls back to cp -a if sourcePath and targetDir are on different
// filesystems.
func link(sourcePath string, targetDir string) error {
	filename := filepath.Base(sourcePath)
	target := filepath.Join(targetDir, filename)
	originalSourcePath := sourcePath

	// make sure we source the underlying file, not a symlink
	// AFTER defining the target to be named as per initial sourcePath
	// This avoids needing to ensure that any link/copy etc. actions
	// follow symlinks and risking hard to find bugs.
	sourcePath, err := filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(targetDir, 0775); err != nil {
		return err
	}

	err = os.Link(sourcePath, target)
	// Error codes: https://man7.org/linux/man-pages/man2/link.2.html
	switch {
	// X-Device link || No permission to link - Try simple copy
	case errors.Is(err, syscall.EXDEV) || errors.Is(err, syscall.EPERM):
		err = copy(sourcePath, target)
	// File already exists - not an err if it's identical
	case errors.Is(err, syscall.EEXIST) && !internal.SameFile(sourcePath, target):
		err = nil
	}

	if err == nil {
		log.Default().Println(originalSourcePath + " -> " + target)
	}

	return err
}

func copy(sourcePath string, target string) error {
	src, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err,
			src.Close(),
		)
	}()

	_srcstat, err := src.Stat()
	if err != nil {
		return err
	}
	dst, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, _srcstat.Mode())
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err,
			dst.Close(),
		)
	}()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err,
			dst.Sync(),
		)
	}()

	err = dst.Chown(int(_srcstat.Sys().(*syscall.Stat_t).Uid), int(_srcstat.Sys().(*syscall.Stat_t).Gid))
	switch {
	// only root has permission (usually)
	case errors.Is(err, syscall.EPERM):
		err = nil
	case err != nil:
		su := int(_srcstat.Sys().(*syscall.Stat_t).Uid)
		sg := int(_srcstat.Sys().(*syscall.Stat_t).Gid)
		_dststat, _ := dst.Stat()
		du := int(_dststat.Sys().(*syscall.Stat_t).Uid)
		dg := int(_dststat.Sys().(*syscall.Stat_t).Gid)
		err = fmt.Errorf("%w (src: %v:%v, dst: %v:%v)", err, su, sg, du, dg)
		return err
	}
	return err
}

// Snaggle parses the file given by path and build minimal /bin & /lib64 under root.
//
// Snaggle will hardlink (or copy, see notes):
//   - path -> root/bin
//   - All dynamically linked dependencies -> root/lib64
//
// For example:
//
//	_ = Snaggle("/bin/which", "/runtime") // you probably want to handle any error, not ignore it
//	// Results in:
//	//  /runtime/bin/which
//	//  /runtime/lib64/libc.so.6
//	//  /runtime/lib64/libpcre2-8.so.0
//	//  /runtime/lib64/libselinux.so.1
//
// # Notes:
//
//   - Future versions intend to provide improved heuristics for destination paths, currently calling
//     Snaggle(path/to/a.library.so) will place a.library.so in root/bin and you need to move it manually
//   - Hardlinks will be created if possible.
//   - A copy will be performed if hardlinking fails for one of the following reasons:
//   - path & root are on different filesystems
//   - the user does not have permission to hardlink (e.g. https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
//   - Copies will retain the original filemode
//   - Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
func Snaggle(path string, root string) error {
	file, err := elf.New(path)
	if err != nil {
		return err
	}
	binDir := filepath.Join(root, "bin")
	libDir := filepath.Join(root, "lib64")
	if err = link(path, binDir); err != nil {
		return err
	}
	for _, lib := range file.Dependencies {
		if err = link(lib, libDir); err != nil {
			return err
		}
	}
	return nil
}
