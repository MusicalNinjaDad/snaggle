package internal

import (
	"path/filepath"
	"regexp"
)

// Regex to check if this is a 64-bit version of `ld-linux*.so`, matches /lib64(/more/directories)/ld-linux*.so(.*)
var Ld_linux_64_RE = regexp.MustCompile(`^\/lib64(?:\/.+|)\/ld-linux.*\.so(?:\..+|)$`)

// Path to interpreter
const P_ld_linux = "/lib64/ld-linux-x86-64.so.2"

// Paths to common libraries
var (
	P_libc       = findLib("libc.so.6")
	P_libm       = findLib("libm.so.6")
	P_libpcre2_8 = findLib("libpcre2-8.so.0")
	P_libpthread = findLib("libpthread.so.0")
	P_libselinux = findLib("libselinux.so.1")
)

// searches /lib* & /usr/lib* to find filename.
func findLib(filename string) (path string) {
	searchPaths := []string{"/lib*/*/", "/usr/lib*/*/", "/lib*/", "/usr/lib*/"} // ld.so-like ordering
	for _, dir := range searchPaths {
		matches, _ := filepath.Glob(dir + filename) // only possible returned error is ErrBadPattern
		if len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}
