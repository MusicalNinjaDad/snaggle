// This is not designed work on non-linux systems - don't try it unless you want to have fun with unexpected
// and unhandled os error types.

package snaggle

import (
	"os"
	"path/filepath"
)

// linkFile creates a hardlink to `path` under `newRoot`, preserving the full directory
// structure similar to how `cp -r` does.
//
// E.g. `linkFile(/usr/bin/which, /tmp)` will create a link at `/tmp/usr/bin/which`.
//
// Note: the _absolute_ `path` will be used, even if a relative path is provided.
//
// Errors may be propogated from `filepath.Abs`, `os.MkdirAll` and `os.Link`, sadly a complete
// lack of meaningful documentation on stdlib errors means the author of this code can't give
// any guidance on what they may be.
// PRs are always welcome to improve error handling or documentation.
func linkFile(path string, newRoot string) (string, error) {
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
