package main

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func parseArgs() {
	// generic
	pflag.String("name", "locker", "Name of container (used in hostname and more)")
	pflag.String("path", linuxDefaultPATH, "Path env variable")

	// cgroups
	pflag.String("memory-limit", "1GB", "RAM limit of container in bytes")
	pflag.Bool("memory-swappiness", false, "Allow swappiness in container (boolean)")
	pflag.String("cpus-allowed", "0", "Number of cpu cores to use in container")

	// network
	pflag.String("network", "forwarding", "Type of network to use")

	// security
	pflag.String("seccomp", "seccomp_default.json", "Seccomp profile path")
	pflag.String("aa-profile-name", "locker-default", "Apparmor profile name")
	pflag.String("aa-template", "locker.prof", "Apparmor profile template")

	pflag.Parse()
}

func bindFlagsToConfig() error {
	// generic
	err := bindFlagToConfig("name", "name")
	if err != nil {
		return err
	}
	err = bindFlagToConfig("path", "path")
	if err != nil {
		return err
	}

	// cgroups
	err = bindFlagToConfig("cgroups.memory-limit", "memory-limit")
	if err != nil {
		return err
	}
	err = bindFlagToConfig("cgroups.memory-swappiness", "memory-swappiness")
	if err != nil {
		return err
	}
	err = bindFlagToConfig("cgroups.cpus-allowed", "cpus-allowed")
	if err != nil {
		return err
	}

	// network
	err = bindFlagToConfig("network.network", "network")
	if err != nil {
		return err
	}

	// security
	err = bindFlagToConfig("security.seccomp", "seccomp")
	if err != nil {
		return err
	}
	err = bindFlagToConfig("security.ap-profile-name", "aa-profile-name")
	if err != nil {
		return err
	}
	err = bindFlagToConfig("security.aa-template", "ap-profile-name")
	if err != nil {
		return err
	}
	return nil
}

func bindFlagToConfig(configName, flagName string) error {
	flag := pflag.Lookup(flagName)
	if flag == nil {
		return fmt.Errorf("flag given for %q is nil", flagName)
	}
	err := viper.BindPFlag(configName, flag)
	if err != nil {
		return errors.Wrapf(err, "Couldn't bind flag %q to %q", flagName, configName)
	}
	return nil
}
