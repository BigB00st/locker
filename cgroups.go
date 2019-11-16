package main

import (
	"os"
	"io/ioutil"
	"strconv"
	"path"
)

//cgroup function, limits recourse usage of process
func cg(config *Config) {
	bytesLimit := 10000000
	swappiness := 0

	os.Mkdir(config.cgroupMemoryPath, 0755)

	//limit RAM
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, byteLimitFile), []byte(strconv.Itoa(bytesLimit)), 0700))

	//disable swapiness
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, swapinessFile), []byte(strconv.Itoa(swappiness)), 0700))
	
	//assign PID to cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, procsFile), []byte(strconv.Itoa(os.Getpid())), 0700))

	//cleanup after container exists
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, notifyOnReleaseFile), []byte("1"), 0700))
}