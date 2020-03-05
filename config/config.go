package config

import (
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitlab.com/amit-yuval/locker/caps"
)

const configFile = "config.toml"
const linuxDefaultPATH = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

func Load() error {
	if err := readConfig(); err != nil {
		return errors.Wrap(err, "couldn't read config file.")
	}
	parseArgs()
	if err := bindFlagsToConfig(); err != nil {
		return errors.Wrap(err, "couldn't bind flags to config")
	}
	if capList, err := caps.GetCapsList(); err != nil {
		return err
	} else {
		viper.Set("security.caps", capList)
	}

	return nil
}

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
	return nil
}
