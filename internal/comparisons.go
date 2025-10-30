package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Are two files identical?, Returns false on any fs/io errors.
func SameFile(path1 string, path2 string) bool {
	same, err := sameFile(path1, path2)
	if err != nil {
		return false
	}
	return same
}

func AssertSameFile(t *testing.T, path1 string, path2 string) {
	t.Helper()

	same, err := sameFile(path1, path2)

	assert.NoError(t, err)
	assert.Truef(t, same, "%s & %s are different files", path1, path2)
}

func AssertLinkedFile(t *testing.T, path1 string, path2 string) {
	t.Helper()

	same, err := sameInode(path1, path2)

	assert.NoError(t, err)
	assert.Truef(t, same, "%s & %s are different files", path1, path2)
}

func sameFile(path1 string, path2 string) (bool, error) {
	same, err := sameInode(path1, path2)
	switch {
	case err != nil:
		return false, err
	case same:
		return true, nil
	case !same:
		// keep checking ...
	}

	same, err = sameModeAndSize(path1, path2)
	switch {
	case err != nil:
		return false, err
	case same:
		// keep checking ...
	case !same:
		return false, nil
	}

	same, err = sameHash(path1, path2)
	switch {
	case err != nil:
		return false, err
	default:
		return same, err
	}
}

// Are two paths refering to identical inodes?
func sameInode(path1 string, path2 string) (bool, error) {
	file1, err1 := os.Stat(path1)
	file2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		err := fmt.Errorf("errors stating: %s - %w; %s - %w", path1, err1, path2, err2)
		return false, err
	}
	return os.SameFile(file1, file2), nil
}

func sameModeAndSize(path1 string, path2 string) (bool, error) {
	file1, err1 := os.Stat(path1)
	file2, err2 := os.Stat(path2)
	if err1 != nil || err2 != nil {
		return false, errors.Join(err1, err2)
	}
	return (file1.Mode() == file2.Mode()) && (file1.Size() == file2.Size()), nil
}

// Do both files have same SHA256?
func sameHash(path1 string, path2 string) (bool, error) {
	sha1, err1 := HashFile(path1)
	sha2, err2 := HashFile(path2)
	return slices.Equal(sha1, sha2), errors.Join(err1, err2)
}

func HashFile(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = errors.Join(err,
			file.Close(),
		)
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func AssertDirectoryContents(t *testing.T, expected []string, dir string) {
	t.Helper()
	a := assert.New(t)
	contents := make([]string, 0, len(expected))
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if !entry.IsDir() {
			contents = append(contents, path)
		}
		return err // lazy - append may have already happened but not important.
	})
	if a.NoError(err) {
		a.ElementsMatch(expected, contents) // expected is coming from a map, so has non-deterministic ordering
	}
}
