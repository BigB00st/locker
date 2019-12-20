package main

import (
	"os"
)

func isChild() bool {
    return os.Getpid() == 1
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}
