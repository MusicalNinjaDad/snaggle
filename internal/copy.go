package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

// Copy a file, maintaining mode & attempting to maintain ownership.
//
// Errors returned will be of type [*fs.PathError] (unless they came from [io.Copy],
// which sadly doesn't document error details ...)
func Copy(sourcePath string, target string, locks *FileLocks) error {
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

	locks.add(target)
	defer locks.remove(target)

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

// We need to lock a file while it is being copied. Othwerwise a second goroutine may attempt to
// create the same link and fail because FileExists && !SameFile
type FileLocks struct {
	m     sync.RWMutex    // to avoid concurrent updates to fileLocks
	locks map[string]bool //keys: paths of locked files
}

func NewFileLock() *FileLocks {
	fl := new(FileLocks)
	fl.locks = make(map[string]bool)
	return fl
}

func (fl *FileLocks) add(path string) {
	fl.m.Lock()
	defer func() { fl.m.Unlock() }()
	fl.locks[path] = true
}

func (fl *FileLocks) wait(path string) {
	fl.m.RLock()
	defer fl.m.RUnlock()
	for {
		if locked, exists := fl.locks[path]; exists && locked {
			continue
		}
		break
	}
}

func (fl *FileLocks) remove(path string) {
	fl.m.Lock()
	defer fl.m.Unlock()
	fl.locks[path] = false
}
