package snaggle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"github.com/u-root/u-root/pkg/ldd"
)

// A symlink or simple file.
//   - Source, Target != "", Err == nil -> Symlink
//   - Source != "", Target == "", Err == nil -> File
//   - Err != nil -> Something else
type Symlink struct {
	Source string
	Target string
	Err    error
}

func LibPaths(bin string) ([]string, error) {
	libs, err := ldd.List(bin)
	slices.Sort(libs)
	return libs, err
}

func NewSymlink(source string) Symlink {
	source, err := filepath.Abs(source)
	if err != nil {
		return Symlink{source, "", err}
	}

	target, err := os.Readlink(source)
	notALink := syscall.EINVAL // https://www.man7.org/linux/man-pages/man2/readlink.2.html#ERRORS
	if errors.Is(err, notALink) {
		return Symlink{source, "", nil}
	}

	return Symlink{source, target, err}
}

func SymlinkChain(path string) []Symlink {
	link := NewSymlink(path)
	chain := []Symlink{link}
	return extendSymlinkChain(chain)
}

func extendSymlinkChain(chain []Symlink) []Symlink {
	lastLink := chain[len(chain)-1]
	if lastLink.Target == "" {
		return chain
	} else {
		nextLink := NewSymlink(lastLink.Target)
		chain = append(chain, nextLink)
		return extendSymlinkChain(chain)
	}
}
