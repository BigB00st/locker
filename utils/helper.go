package utils

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
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

// function recieves an array of FOO=bar envs and sets
func SetEnv(envList []string) error {
	for _, env := range envList {
		split := strings.SplitN(env, "=", 2)
		if len(split) != 2 {
			return errors.Errorf("env {%s} invalid", env)
		}
		if err := os.Setenv(split[0], split[1]); err != nil {
			return errors.Wrapf(err, "couldn't set env {%s}", env)
		}
	}
	return nil
}

// function gets interface array and return string array
func InterfaceArrToStrArr(arr []interface{}) []string {
	ret := make([]string, len(arr))
	for i, v := range arr {
		ret[i] = v.(string)
	}
	return ret
}
