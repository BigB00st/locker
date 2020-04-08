package utils

import (
	"os"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
)

// GetFdFromPath returns file descriptor from path
func GetFdFromPath(path string) (int, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return -1, err
	}
	return fd, nil
}

// ResolvePath returns full path if exists (resolving link if necessary)
func resolvePath(path, baseDir string, envList []string) (string, error) {
	fullPath := filepath.Join(baseDir, path)
	if FileExists(fullPath) {
		return fullPath, nil
	} else if LinkExists(fullPath) {
		path, _ = os.Readlink(fullPath)
		return filepath.Join(baseDir, path), nil
	}
	return "", errors.Errorf("executable %s doesn't exist", path)
}

// FileExists true if file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// LinkExists true if file exists
func LinkExists(path string) bool {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
