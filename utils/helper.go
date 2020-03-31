package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	uuid "github.com/nu7hatch/gouuid"
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
		retList = append(retList, strings.Split(split[1], ":")...)
	}
	return retList, nil
}

func GetExecutablePath(executable, basePath string, envList []string) (string, error) {
	PATH, err := getPATHList(envList)
	if err != nil {
		return "", err
	}
	if strings.Contains(executable, "/") { //absolute path
		curPath := filepath.Join(basePath, executable)
		if Exists(curPath) {
			return curPath, nil
		}
	} else { //relative path, loop over PATH to find
		for _, v := range PATH {
			curPath := filepath.Join(basePath, v, executable)
			if Exists(curPath) {
				return curPath, nil
			}
		}
	}

	return "", errors.Errorf("couldn't find executable %s", executable)
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
	a = append(a[:i], a[i+1:]...)
}

func findElement(element string, arr []string) int {
	for i := range arr {
		if arr[i] == element {
			return i
		}
	}
	return -1
}

type createFunc func(length int) (string, error)
type isUniqueFunc func(arg string) bool

func GetUnique(prefix string, length int, create createFunc, isUnique isUniqueFunc) (string, error) {
	var ret string
	createLen := length - len(prefix)
	for {
		curStr, err := create(createLen)
		if err != nil {
			return "", err
		}
		ret = prefix + curStr
		if !isUnique(ret) {
			break
		}
	}
	return ret, nil
}

func CreateUuid(length int) (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrap(err, "couldn't create uuid")
	}
	return u.String()[:length], nil
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}

func PadSpaces(pad int, list ...string) string {
	var ret string
	for _, v := range list {
		ret += v + strings.Repeat(" ", pad-len(v))
	}
	return ret
}
