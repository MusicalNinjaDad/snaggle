package symlink

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

const notALink = syscall.EINVAL // https://www.man7.org/linux/man-pages/man2/readlink.2.html#ERRORS

// A symlink or simple file.
//   - Source, Target != "", Err == nil -> Symlink
//   - Source != "", Target == "", Err == nil -> File
//   - Err != nil -> Something else
type Symlink struct {
	Source string
	Target string
	Err    error
}

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

func Chain(path string) []Symlink {
	link := New(path)
	chain := []Symlink{link}
	return extendSymlinkChain(chain)
}

func extendSymlinkChain(chain []Symlink) []Symlink {
	lastLink := chain[len(chain)-1]
	if lastLink.Target == "" {
		return chain
	} else {
		nextLink := New(lastLink.Target)
		chain = append(chain, nextLink)
		return extendSymlinkChain(chain)
	}
}
