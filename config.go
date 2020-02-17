package main

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitlab.com/bigboost/locker/cgroups"
)

func readConfig() error {
	viper.SetConfigName(configFile) // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/locker/")
	viper.AddConfigPath(".") // Checks in current directory, Only for debugging purposes
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok { //ignore config file not found
			return errors.Wrapf(err, "Error while reading %q config file", configFile)
		}
	}
	viper.Set("cgroups.name", "locker"+strconv.Itoa(os.Getpid()))
	viper.Set("cgroups.cpuset-path", path.Join(cgroups.BasePath, cgroups.CPUSetPath, viper.GetString("cgroups.name")))
	viper.Set("cgroups.cpuset-root-path", path.Join(cgroups.BasePath, cgroups.CPUSetPath))
	viper.Set("cgroups.memory-path", path.Join(cgroups.BasePath, cgroups.MemoryPath, viper.GetString("cgroups.name")))
	viper.Set("cgroups.memory-root-path", path.Join(cgroups.BasePath, cgroups.MemoryPath))
	return nil
}
