package utils

import (
	"os"
	"syscall"
)

// returns file descriptor from path
func GetFdFromPath(path string) (int, error) {
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return -1, err
	}
	return fd, nil
}

// function returns true if file/link/dir exists
func Exists(path string) bool {
	return FileExists(path) || LinkExists(path)
}

// returns true if file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// returns true if file exists
func LinkExists(path string) bool {
	if _, err := os.Lstat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}
