package internal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
)

// Copy a file, maintaining mode & attempting to maintain ownership.
//
// Errors returned will be of type [*fs.PathError] (unless they came from [io.Copy],
// which sadly doesn't document error details ...)
func Copy(sourcePath string, target string) error {
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
