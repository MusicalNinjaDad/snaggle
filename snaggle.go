// Snag a copy of a ELF binary and all its dependencies to another/path/bin & another/path//lib64.
//
// This is the main implementation of the command-line application `snaggle`, for use as a library in
// other code and scripts.
//
// Snaggle is designed to help create minimal runtime containers from pre-existing installations.
// It may work for other use cases and I'd be interested to hear about them at:
// https://github.com/MusicalNinjaDad/snaggle
//
// WARNING: This is not designed work on non-linux systems - don't try it unless you want to have fun
// with unexpected and unhandled os error types.
package snaggle

import (
	debug_elf "debug/elf"
	"errors"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/MusicalNinjaDad/snaggle/elf"
	"github.com/MusicalNinjaDad/snaggle/internal"
)

func init() {
	log.SetFlags(0)
}

// create a hardlink in targetDir which references sourcePath,
// falls back to cp -a if sourcePath and targetDir are on different
// filesystems.
//
// Errors returned will be [*fs.PathError] with Path=sourcePath,
// errors wrapped further down the chain may include the resolved path.
// Our PathError may wrap a further PathError or an [*os.LinkError].
func link(sourcePath string, targetDir string, checker chan<- skipCheck) (err error) {
	originalSourcePath := sourcePath

	op := "resolve target"
	targetDir, err = filepath.Abs(targetDir)
	if err != nil {
		return &fs.PathError{Op: op, Path: originalSourcePath, Err: err}
	}
	filename := filepath.Base(sourcePath)
	target := filepath.Join(targetDir, filename)

	// make sure we source the underlying file, not a symlink
	// AFTER defining the target to be named as per initial sourcePath
	// This avoids needing to ensure that any link/copy etc. actions
	// follow symlinks and risking hard to find bugs.
	op = "resolve"
	sourcePath, err = filepath.EvalSymlinks(sourcePath)
	if err != nil {
		return &fs.PathError{Op: op, Path: originalSourcePath, Err: err}
	}

	op = "mkdir target"
	if err := os.MkdirAll(targetDir, 0775); err != nil {
		return &fs.PathError{Op: op, Path: originalSourcePath, Err: err}
	}

	reply := make(chan bool)
	checker <- skipCheck{target, reply}

	if <-reply {
		op = "skip"
	} else {
		op = "link"
		err = os.Link(sourcePath, target)
		// Error codes: https://man7.org/linux/man-pages/man2/link.2.html
		switch {
		// X-Device link || No permission to link - Try simple copy
		case errors.Is(err, syscall.EXDEV) || errors.Is(err, syscall.EPERM):
			op = "copy"
			err = internal.Copy(sourcePath, target)
		// File already exists - not an err if it's identical
		case errors.Is(err, syscall.EEXIST) && internal.SameFile(sourcePath, target):
			err = nil
		}
	}

	if err != nil {
		return &fs.PathError{Op: op, Path: originalSourcePath, Err: err}
	}
	if originalSourcePath == sourcePath {
		log.Default().Println(op + " " + originalSourcePath + " -> " + target)
	} else {
		log.Default().Println(op + " " + originalSourcePath + " (" + sourcePath + ") -> " + target)
	}
	return nil

}

// Snaggle parses the file(s) given by path and build minimal /bin & /lib64 under root.
//
// If path refers to a directory, all valid ELF binaries directly under path will be snagged.
// Provide the Option [Recursive()] to recurse subdirectories.
//
// Snaggle will hardlink (or copy, see notes):
//   - path -> root/bin (executables) or path/lib64 (libraries), unless the Option [InPlace()] is provided
//   - All dynamically linked dependencies -> root/lib64
//
// For example:
//
//	_ = Snaggle("/bin/which", "/runtime") // you probably want to handle any error, not ignore it
//	// Results in:
//	//  /runtime/bin/which
//	//  /runtime/lib64/libc.so.6
//	//  /runtime/lib64/libpcre2-8.so.0
//	//  /runtime/lib64/libselinux.so.1
//
// # Notes:
//
//   - Hardlinks will be created if possible.
//   - A copy will be performed if hardlinking fails for one of the following reasons:
//   - path & root are on different filesystems
//   - the user does not have permission to hardlink (e.g. https://docs.kernel.org/admin-guide/sysctl/fs.html#protected-hardlinks)
//   - Copies will retain the original filemode
//   - Copies will attempt to retain the original ownership, although this will likely fail if running as non-root
func Snaggle(path string, root string, opts ...Option) error {
	snaggerrs := new(errgroup.Group)

	checker := make(chan skipCheck)
	go skipHandler(checker)

	var options options
	for _, optfn := range opts {
		optfn(&options)
	}

	switch {
	case options.copy && options.inplace:
		return &InvocationError{Path: path, Target: root, err: ErrCopyInplace}
	case !options.verbose:
		output := log.Writer()
		log.SetOutput(io.Discard)
		defer log.SetOutput(output)
	case options.verbose:
		snaggerrs.SetLimit(1)
	}

	snagfile := func(path string) error {
		var badelf *debug_elf.FormatError
		err := snaggle(path, root, options, checker)
		switch {
		case err == nil:
			return nil // snagged
		case errors.As(err, &badelf):
			return nil // not an ELF
		default:
			return err
		}
	}

	var snagdir func(dir string) error //see https://github.com/golang/go/issues/226 :-x FFS!
	snagdir = func(dir string) error {
		files, err := os.ReadDir(dir)
		if err != nil {
			return err
		}

		for _, file := range files {
			path := filepath.Join(dir, file.Name())
			filemode := file.Type()

			isDir := filemode&fs.ModeDir != 0
			if filemode&fs.ModeSymlink != 0 {
				p, _ := filepath.EvalSymlinks(path) //known valid path & symlink
				s, err := os.Stat(p)
				if err != nil {
					return err
				}
				isDir = s.IsDir()
			}

			switch {
			case isDir && options.recursive:
				if err := snagdir(path); err != nil {
					return err
				}
			case isDir:
				continue // skip Directory entries
			default:
				snaggerrs.Go(func() error { return snagfile(path) })
			}
		}

		return nil
	}

	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	switch {
	case stat.IsDir():
		if err := snagdir(path); err != nil {
			return &SnaggleError{Src: path, Dst: root, err: err}
		}
		return snaggerrs.Wait()
	case options.recursive:
		err = &fs.PathError{Op: "--recursive", Path: path, Err: syscall.ENOTDIR}
		return &InvocationError{Path: path, Target: root, err: err}
	default:
		return snaggle(path, root, options, checker)
	}
}

func snaggle(path string, root string, options options, checker chan<- skipCheck) error {
	binDir := filepath.Join(root, "bin")
	libDir := filepath.Join(root, "lib64")
	file, err := elf.New(path)
	switch {
	case err == nil:
		break
	case options.copy:
		var formatError *debug_elf.FormatError
		if errors.As(err, &formatError) {
			break
		} else {
			return &SnaggleError{path, "", err}
		}
	default:
		return &SnaggleError{path, "", err}
	}

	linkerrs := new(errgroup.Group)

	switch {
	case options.verbose:
		linkerrs.SetLimit(1)
	}

	switch {
	case options.inplace:
		// do not link file
	case options.copy:
		dest := filepath.Join(root, filepath.Dir(path))
		linkerrs.Go(func() error { return link(path, dest, checker) })
	default:
		if file.IsExe() {
			linkerrs.Go(func() error { return link(path, binDir, checker) })
		} else {
			linkerrs.Go(func() error { return link(path, libDir, checker) })
		}
	}

	// TODO: #50 make linking interpreter safer
	if file.Interpreter != "" {
		linkerrs.Go(func() error { return link(file.Interpreter, libDir, checker) }) // currently OK - as it sits in /lib64 ... but ...
	}

	for _, lib := range file.Dependencies {
		linkerrs.Go(func() error { return link(lib, libDir, checker) })
	}

	// TODO: #37 improve error handling with context, error collector, rollback
	//       (probably requires link to return path of file created, if created)
	if err := linkerrs.Wait(); err != nil {
		return &SnaggleError{Src: path, Dst: root, err: err}
	}
	return nil
}

// options used by [Snaggle]
type options struct {
	copy      bool // copy entire directory contents to /destinationroot/full/source/path
	inplace   bool // snag in place, only snag dependencies & interpreter
	recursive bool // recurse subdirectories & snag everything
	verbose   bool // output to stdout and process sequentially for readability
}

// Option setting functions
type Option func(*options)

// Copy entire directory contents to /destinationroot/full/source/path
func Copy() Option { return func(o *options) { o.copy = true } }

// Snag in place: only snag dependencies & interpreter
func InPlace() Option { return func(o *options) { o.inplace = true } }

// Snag recursively: only works when snaggling a directory
func Recursive() Option { return func(o *options) { o.recursive = true } }

// Output to stdout and process sequentially for readability
func Verbose() Option { return func(o *options) { o.verbose = true } }

// An error occurred during snaglling
type SnaggleError struct {
	Src string // Source path
	Dst string // Destination path
	err error
}

func (e *SnaggleError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *SnaggleError) Unwrap() error {
	return e.err
}

// Snaggle was invoked with semantically invalid inputs.
// err will be formatted to read well when directly output to stderr during CLI invokation
// E.g: invoking Snaggle(path/to/FILE, root, Recursive()) will wrap a
// [&fs.PathError]{Op: "--recursive", Path: "path/to/FILE", Err: syscall.ENOTDIR}
type InvocationError struct {
	Path   string
	Target string
	err    error
}

var (
	ErrCopyInplace = errors.New("cannot copy in-place")
)

func (e *InvocationError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (e *InvocationError) Unwrap() error {
	return e.err
}

type skipCheck struct {
	destination string
	response    chan bool
}

func skipHandler(in <-chan skipCheck) {
	linked := make(map[string]bool)
	for {
		request := <-in
		_, exists := linked[request.destination]
		if !exists {
			linked[request.destination] = true
			request.response <- false
			continue
		}
		request.response <- true
	}
}
