package main

import (
	"syscall"
	"fmt"
)

// returns file descriptor from path
func getFdFromPath(path string) (int, error) {
	fmt.Println("opening: ", path)
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return -1, err
	}
	return fd, nil
}