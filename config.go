package main

import (
	"path"
	"strconv"
	"os"
)

type Config struct {
	name string
	cgroupMemoryPath string
	cgroupCPUPath string
}

func NewConfig() *Config {
	config := new(Config)
	config.name = "locker" + strconv.Itoa(os.Getpid())
	config.cgroupCPUPath = path.Join(cgroupPath, cgroupCPU, config.name)
	config.cgroupMemoryPath = path.Join(cgroupPath, cgroupMemory, config.name)
	
	return config
}