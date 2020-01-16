package main

import (
    "os"
    "strings"
)

func isChild() bool {
    return os.Getpid() == 1
}

func stringInSlice(a string, list []string) bool {
    for _, b := range list {
        if strings.Contains(b, a) {
            return true
        }
    }
    return false
}
