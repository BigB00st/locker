package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// function recieves an array of PATH=/path envs and returns list of paths only
func getPATHList(envList []string) ([]string, error) {
	var retList []string
	for _, env := range envList {
		split := strings.SplitN(env, "=", 2)
		if len(split) != 2 {
			return nil, errors.Errorf("env {%s} invalid", env)
		}
		retList = append(retList, split[1])
	}
	return retList, nil
}

func GetExecutablePath(executable, basePath string, envList []string) (string, error) {
	PATH, err := getPATHList(envList)
	if err != nil {
		return "", err
	}
	for _, v := range PATH {
		curPath := filepath.Join(basePath, v)
		if fileExists(curPath) {
			return curPath, nil
		}
	}

	return "", errors.New("couldn't find executable %s")
}

// returns true if file exists
func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// function gets interface array and return string array
func InterfaceArrToStrArr(arr []interface{}) []string {
	ret := make([]string, len(arr))
	for i, v := range arr {
		ret[i] = v.(string)
	}
	return ret
}

func GetChildArgs(cmdList []string) []string {
	ret := os.Args[1:]            //remove "locker"
	ret = append(ret, cmdList...) // add cmdList
	deleteElement("run", ret)     //remove "run"
	return ret[:len(ret)-1]       //last item is duplicated for some reason, maybe bug in go?
}

func deleteElement(element string, a []string) {
	i := findElement(element, a)
	fmt.Println("BEFORE:", a)
	a = append(a[:i], a[i+1:]...)
}

func findElement(element string, arr []string) int {
	fmt.Println("SEARCHING", arr, "FOR", element)
	for i := range arr {
		if arr[i] == element {
			fmt.Println("found", element, "index:", i)
			return i
		}
	}
	return -1
}
