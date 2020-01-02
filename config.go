package main

import (
	"path"
	"strconv"
)

type Config struct {
	name string
	pid int
	cgroupMemoryPath string
	cgroupCPUSetPath string
	cgroupCPUSetRootPath string
}

func NewConfig(pid int) *Config {
	config := new(Config)
	config.pid = pid
	config.name = "locker" + strconv.Itoa(pid)
	config.cgroupCPUSetPath = path.Join(cgroupPath, cgroupCPUSet, config.name)
	config.cgroupCPUSetRootPath = path.Join(cgroupPath, cgroupCPUSet)
	config.cgroupMemoryPath = path.Join(cgroupPath, cgroupMemory, config.name)

	return config
}
