package io

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sys/unix"
)

// function runs command and return output as string
func CmdOut(binary string, arg ...string) (string, error) {
	c := exec.Command(binary, arg...)

	output, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running `%s %s` failed with output: %s\nerror: %v", c.Path, strings.Join(c.Args, " "), output, err)
	}

	return string(output), nil
}

// GetFdFromPath returns file descriptor from path
func GetFdFromPath(path string) (int, error) {
	fd, err := unix.Open(path, unix.O_RDONLY, 0)
	if err != nil {
		return -1, err
	}
	return fd, nil
}

// ResolvePath returns full path if exists (resolving link if necessary)
func ResolvePath(path, baseDir string) (string, error) {
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

// Copy copies dst to src, returns number of bytes written
func Copy(src, dst string) (int64, error) {
	srcFileFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !srcFileFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer dstFile.Close()

	nBytes, err := io.Copy(dstFile, srcFile)
	return nBytes, err
}

// MkdirIfNotExist creates given directory if it doesn't exist
func MkdirIfNotExist(dir string) {
	if !FileExists(dir) {
		os.MkdirAll(dir, os.ModeDir)
	}
}
