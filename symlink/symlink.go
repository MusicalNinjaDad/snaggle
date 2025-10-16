package symlink

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

const notALink = syscall.EINVAL // https://www.man7.org/linux/man-pages/man2/readlink.2.html#ERRORS

// A symlink (or other filesystem entry - use `Symlink.IsNotLink()` to identify these cases)
//
// Do not construct this manually, use `symlink.New()` or you risk breaking semantics
type Symlink struct {
	Source string // The source path - always absolute
	Target string // Target path (`""` if it's not a link, and maybe also on error) - can be relative
	Err    error  // If we got an error when trying to identify the target
}

// Accurately identify when a Symlink Source is not _actually_ refering to a symlink
func (s Symlink) IsNotLink() bool {
	return s.Err == nil && s.Target == ""
}

// Construct a Symlink from a source path. If a relative path is provided for `source` it will be
// converted to an absolute path.
func New(source string) Symlink {
	source, err := filepath.Abs(source)
	if err != nil {
		return Symlink{source, "", err}
	}
	target, err := os.Readlink(source)
	if errors.Is(err, notALink) {
		return Symlink{source, "", nil}
	}

	return Symlink{source, target, err}
}

// Chain follows a sequence of symbolic links starting at path and returns
// the full chain as a slice of Symlink values.
//
// - The returned slice always contains at least one element for the initial path.
// - If a link cannot be resolved, Chain appends that Symlink with its Err set
// and returns the partial chain along with the error.
// - When Chain reaches a path which is not symlink, the chain ends with a Symlink
// whose Target is "" and Err is nil.
func Chain(path string) ([]Symlink, error) {
	source := path
	var chain []Symlink
	for {
		link := New(source)
		chain = append(chain, link)
		if link.Err != nil {
			// Something went wrong
			return chain, link.Err
		}
		if link.IsNotLink() {
			// We're done here
			return chain, nil
		}
		// keep going with the next link in the chain ...
		if filepath.IsAbs(link.Target) {
			source = link.Target
		} else {
			sourcedir := filepath.Dir(link.Source)
			source = filepath.Join(sourcedir, link.Target)
		}
	}
}
