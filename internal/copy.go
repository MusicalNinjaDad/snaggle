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

	locks.lock(target)
	defer locks.unlock(target)

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
	m     sync.RWMutex             // to avoid concurrent updates to fileLocks
	locks map[string]*sync.RWMutex //keys: paths of locked files
}

func NewFileLock() *FileLocks {
	fl := new(FileLocks)
	fl.locks = make(map[string]*sync.RWMutex)
	return fl
}

func (fl *FileLocks) lock(path string) {
	println("locking lockbox for lock " + path)
	fl.m.Lock()
	println("locking " + path)
	if _, exists := fl.locks[path]; !exists {
		fl.locks[path] = new(sync.RWMutex)
	}
	fl.locks[path].Lock()
	println("unlocking lockbox for lock " + path)
	fl.m.Unlock()
}

func (fl *FileLocks) rlock(path string) {
	println("Rlocking lockbox for rlock" + path)
	fl.m.RLock()
	if lock, exists := fl.locks[path]; exists {
		println("unRlocking lockbox for rlock (known)" + path)
		fl.m.RUnlock()
		println("rlocking " + path)
		lock.RLock()
	} else {
		println("unRlocking lockbox for rlock(unknown)" + path)
		fl.m.RUnlock()
	}
}

func (fl *FileLocks) runlock(path string) {
	println("Rlocking lockbox for runlock " + path)
	fl.m.RLock()
	if lock, exists := fl.locks[path]; exists {
		println("unRlocking lockbox for runlock (known)" + path)
		fl.m.RUnlock()
		println("runlocking " + path)
		lock.RUnlock()
	}
}

func (fl *FileLocks) unlock(path string) {
	println("unlocking " + path)
	fl.locks[path].Unlock()
}
