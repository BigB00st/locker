package config

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

	// capabilites
	pflag.StringSlice("cap-add", nil, "Add linux capabilites")
	pflag.StringSlice("cap-drop", nil, "Drop linux capabilites")

	pflag.Parse()
}

func bindFlagsToConfig() error {
	// generic
	if err := bindFlagToConfig("name", "name"); err != nil {
		return err
	}
	if err := bindFlagToConfig("path", "path"); err != nil {
		return err
	}

	// cgroups
	if err := bindFlagToConfig("cgroups.memory-limit", "memory-limit"); err != nil {
		return err
	}
	if err := bindFlagToConfig("cgroups.memory-swappiness", "memory-swappiness"); err != nil {
		return err
	}
	if err := bindFlagToConfig("cgroups.cpus-allowed", "cpus-allowed"); err != nil {
		return err
	}

	// network
	if err := bindFlagToConfig("network.network", "network"); err != nil {
		return err
	}

	// security
	if err := bindFlagToConfig("security.seccomp", "seccomp"); err != nil {
		return err
	}
	if err := bindFlagToConfig("security.aa-profile-name", "aa-profile-name"); err != nil {
		return err
	}
	if err := bindFlagToConfig("security.aa-template", "aa-profile-name"); err != nil {
		return err
	}
	if err := bindFlagToConfig("security.cap-add", "cap-add"); err != nil {
		return err
	}
	if err := bindFlagToConfig("security.cap-drop", "cap-drop"); err != nil {
		return err
	}
	return nil
}

func bindFlagToConfig(configName, flagName string) error {
	flag := pflag.Lookup(flagName)
	if flag == nil {
		return fmt.Errorf("flag given for %q is nil", flagName)
	}
	if err := viper.BindPFlag(configName, flag); err != nil {
		return errors.Wrapf(err, "Couldn't bind flag %q to %q", flagName, configName)
	}
	return nil
}
