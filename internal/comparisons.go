package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
)

// Are two files identical?, Returns false on any fs/io errors.
// Will wait until dest is not locked before proceeding.
func SameFile(src string, dest string) bool {
	same, err := sameFile(src, dest)
	if err != nil {
		return false
	}
	return same
}

func sameFile(path1 string, path2 string) (bool, error) {
	same, err := SameInode(path1, path2)
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
func SameInode(path1 string, path2 string) (bool, error) {
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
