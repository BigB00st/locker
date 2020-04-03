package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/amit-yuval/locker/caps"
)

// Init reads args, calls setModifiedFlags
func Init() error {
	parseArgs()
	viper.BindPFlags(pflag.CommandLine)
	if err := setModifiedFlags(); err != nil {
		return err
	}

	return nil
}

// setModifiedFlags sets flags at runtime
func setModifiedFlags() error {
	if capList, err := caps.GetCapsList(); err != nil {
		return err
	} else {
		viper.Set("caps", capList)
	}
	return nil
}
