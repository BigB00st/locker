package main

import (
	"path"
	"strconv"
	"os"
)

type Config struct {
	name string
	cgroupMemoryPath string
	cgroupMemoryRootPath string
	cgroupCPUSetPath string
	cgroupCPUSetRootPath string
}

func NewConfig() *Config {
	config := new(Config)
	config.name = "locker" + strconv.Itoa(os.Getpid())
	config.cgroupCPUSetPath = path.Join(cgroupPath, cgroupCPUSet, config.name)
	config.cgroupCPUSetRootPath = path.Join(cgroupPath, cgroupCPUSet)
	config.cgroupMemoryPath = path.Join(cgroupPath, cgroupMemory, config.name)
	config.cgroupMemoryRootPath = path.Join(cgroupPath, cgroupMemory)

	return config
}
