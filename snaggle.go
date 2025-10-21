// This is not designed work on non-linux systems - don't try it unless you want to have fun with unexpected
// and unhandled os error types.

package snaggle

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/MusicalNinjaDad/snaggle/elf"
)

// linkTree creates a hardlink to `path` under `newRoot`, preserving the full directory
// structure similar to how `cp -r` does.
//
// E.g. `linkTree(/usr/bin/which, /tmp)` will create a link at `/tmp/usr/bin/which`.
//
// Note: the _absolute_ `path` will be used, even if a relative path is provided.
//
// Errors may be propogated from `filepath.Abs`, `os.MkdirAll` and `os.Link`, sadly a complete
// lack of meaningful documentation on stdlib errors means the author of this code can't give
// any guidance on what they may be.
// PRs are always welcome to improve error handling or documentation.
func linkTree(path string, newRoot string) (string, error) {
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

// create a hardlink in targetDir which references sourcePath,
// falls back to cp -a if sourcePath and targetDir are on different
// filesystems.
func link(sourcePath string, targetDir string) error {
	filename := filepath.Base(sourcePath)
	target := filepath.Join(targetDir, filename)

	// check target is 404

	if err := os.MkdirAll(targetDir, 0775); err != nil {
		return err
	}

	// TODO: what if either is a symlink?
	err := os.Link(sourcePath, target)
	if errors.Is(err, syscall.EXDEV) || errors.Is(err, syscall.EPERM) {
		// X-Device link || No permission to link
		// Try simple copy
		if err := copy(sourcePath, target); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
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

	dst, err := os.Create(target)
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

	return nil
}

// Parse file and build a minimal /bin & /lib under root
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
