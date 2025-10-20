package internal

import (
	"path/filepath"
	"regexp"
	"strings"
)

// If both paths are absolute: compares only the filename, otherwise compares the entire path.
func Libpathcmp(path1 string, path2 string) int {
	if filepath.IsAbs(path1) && filepath.IsAbs(path2) {
		return strings.Compare(filepath.Base(path1), filepath.Base(path2))
	}
	return strings.Compare(path1, path2)
}

// Regex to check if this is a 64-bit version of `ld-linux*.so`, matches /lib64(/more/directories)/ld-linux*.so(.*)
var Ld_linux_64_RE = regexp.MustCompile(`^\/lib64(?:\/.+|)\/ld-linux.*\.so(?:\..+|)$`)
