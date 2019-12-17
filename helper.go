package main

import (
	"os"
)

func isChild() bool {
    return os.Getpid() == 1
}
