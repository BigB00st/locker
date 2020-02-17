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

// returns true if file exists
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
