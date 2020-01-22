package main

import (
	"github.com/spf13/viper"
	"fmt"
)

func ReadConfig() {
	viper.SetConfigName("config")          // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/locker/")
	viper.AddConfigPath(".")               // Checks in current directory, Only for debugging purposes
	must(viper.ReadInConfig())
	//fmt.Println("container name = ", viper.Get("name"))
}
