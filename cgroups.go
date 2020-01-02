package main

import (
	"os"
	"io/ioutil"
	"strconv"
	"path"
)

//cgroup function, limits recourse usage of process
func (config *Config) CgInit() {
	bytesLimit := 10000000
	swappiness := 0
	cpusAllowed := "0" //Only allow the first cpu

	//make cgruops
	os.Mkdir(config.cgroupMemoryPath, 0755)
	os.Mkdir(config.cgroupCPUSetPath, 0755)

	//limit RAM
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, byteLimitFile), []byte(strconv.Itoa(bytesLimit)), 0700))

	//disable swapiness
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, swapinessFile), []byte(strconv.Itoa(swappiness)), 0700))

	//limit amount of CPUs allowes
	mems, readErr := ioutil.ReadFile(path.Join(config.cgroupCPUSetRootPath, cpusetMemFile))
	must(readErr)
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, cpusetMemFile), mems, 0700))
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, cpusetLimitFile), []byte(cpusAllowed), 0700))

	//assign PID to memory cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, procsFile), []byte(strconv.Itoa(config.pid)), 0700))

	//assign PID to cpuset cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, procsFile), []byte(strconv.Itoa(config.pid)), 0700))

	//cleanup after container exists
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, notifyOnReleaseFile), []byte("1"), 0700))
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, notifyOnReleaseFile), []byte("1"), 0700))
}

//cgroup function, limits recourse usage of process
func (config *Config) CgDestruct() {
	must(os.Remove(config.cgroupMemoryPath))
	must(os.Remove(config.cgroupCPUSetPath))
}
