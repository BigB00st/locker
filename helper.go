package main

import (
	"os"
	"strings"
)

// Child is created in a new pid namespace, so it will gain pid 1
func isChild() bool {
	return os.Getpid() == 1
}

// Function returns true if given string is in given list of strings
func stringInSlice(str string, list []string) bool {
	for _, curStr := range list {
		if strings.Contains(curStr, str) {
			return true
		}
	}
	return false
}
