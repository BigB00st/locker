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
	mems, err := ioutil.ReadFile(path.Join(config.cgroupCPUSetRootPath, cpusetMemFile))
	must(err)
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, cpusetMemFile), mems, 0700))
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, cpusetLimitFile), []byte(cpusAllowed), 0700))

	//assign self to memory cgroup

	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, procsFile), []byte("0"), 0700))

	//assign self to cpuset cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, procsFile), []byte("0"), 0700))

	//cleanup after container exists
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryPath, notifyOnReleaseFile), []byte("1"), 0700))
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetPath, notifyOnReleaseFile), []byte("1"), 0700))
}

func CgRemoveSelf(config *Config) {
	//assign self to root memory cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupMemoryRootPath, procsFile), []byte("0"), 0700))

	//assign self to root cpuset cgroup
	must(ioutil.WriteFile(path.Join(config.cgroupCPUSetRootPath, procsFile), []byte("0"), 0700))
}

//cgroup function, limits recourse usage of process
func (config *Config) CgDestruct() {
	must(os.Remove(config.cgroupMemoryPath))
	must(os.Remove(config.cgroupCPUSetPath))
}
