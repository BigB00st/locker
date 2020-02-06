package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func parseArgs() {
	// generic
	pflag.String("name", "locker", "Name of container (used in hostname and more)")
	pflag.String("path", "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "Path env variable")

	// cgroups
	pflag.String("memory-limit", "1GB", "RAM limit of container in bytes")
	pflag.Bool("memory-swappiness", false, "Allow swappiness in container (boolean)")
	pflag.String("cpus-allowed", "0", "Number of cpu cores to use in container")

	// network
	pflag.String("network", "forwarding", "Type of network to use")

	// security
	pflag.String("seccomp", "seccomp_default.json", "Seccomp profile name")
	pflag.String("apparmor", "not-implemented", "not-implemented")

	pflag.Parse()
}

func bindFlagsToConfig() {
	// generic
	viper.BindPFlag("name", pflag.Lookup("name"))
	viper.BindPFlag("path", pflag.Lookup("path"))

	// cgroups
	viper.BindPFlag("cgroups.memory-limit", pflag.Lookup("memory-limit"))
	viper.BindPFlag("cgroups.memory-swappiness", pflag.Lookup("memory-swappiness"))
	viper.BindPFlag("cgroups.cpus-allowed", pflag.Lookup("cpus-allowed"))

	// network
	viper.BindPFlag("network.network", pflag.Lookup("network"))

	// security
	viper.BindPFlag("security.seccomp", pflag.Lookup("seccomp"))
	viper.BindPFlag("security.apparmor", pflag.Lookup("apparmor"))
}
