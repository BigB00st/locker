package main

import (
	"path"
	"strconv"
)

type Config struct {
	name string
	pid int
	cgroupMemoryPath string
	cgroupCPUPath string
}

func NewConfig(pid int) *Config {
	config := new(Config)
	config.pid = pid
	config.name = "locker" + strconv.Itoa(pid)
	config.cgroupCPUPath = path.Join(cgroupPath, cgroupCPU, config.name)
	config.cgroupMemoryPath = path.Join(cgroupPath, cgroupMemory, config.name)
	
	return config
}