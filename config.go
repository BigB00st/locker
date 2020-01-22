package main

import (
	"github.com/spf13/viper"
	"path"
	"strconv"
	"os"
)

func readConfig() {
	viper.SetConfigName("config")          // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/locker/")
	viper.AddConfigPath(".")               // Checks in current directory, Only for debugging purposes
	must(viper.ReadInConfig())
	viper.Set("cgroups.name", "locker" + strconv.Itoa(os.Getpid()))
	viper.Set("cgroups.cpuset-path", path.Join(cgroupPath, cgroupCPUSet, viper.GetString("cgroups.name")))
	viper.Set("cgroups.cpuset-root-path", path.Join(cgroupPath, cgroupCPUSet))
	viper.Set("cgroups.memory-path", path.Join(cgroupPath, cgroupMemory, viper.GetString("cgroups.name")))
	viper.Set("cgroups.memory-root-path", path.Join(cgroupPath, cgroupMemory))
	//fmt.Println("container name = ", viper.Get("name"))
}
