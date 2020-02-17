package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Child is created in a new pid namespace, so it will gain pid 1
func IsChild() bool {
	return os.Getpid() == 1
}

// Function returns true if given string is in given list of strings
func StringInSlice(str string, list []string) bool {
	for _, curStr := range list {
		if strings.Contains(curStr, str) {
			return true
		}
	}
	return false
}

// function runs command and return output as string
func CmdOut(binary string, arg ...string) (string, error) {
	c := exec.Command(binary, arg...)

	output, err := c.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("running `%s %s` failed with output: %s\nerror: %v", c.Path, strings.Join(c.Args, " "), output, err)
	}

	return string(output), nil
}

func PrintAndExit(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}
