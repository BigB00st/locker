package main

import (
	"github.com/spf13/pflag"
)

func parseArgs() {

	// genereic
	pflag.String("name", "locker", "Name of container (used in hostname and more)")

	// cgroups
	pflag.String("memory-limit", "1GB", "RAM limit of container in bytes")
	pflag.Bool("memory-swappiness", false, "Allow swappiness in container (boolean)")
	pflag.Int("cpu-count", 1, "Number of cpu cores to use in container")

	// network
	pflag.String("network", "forwarding", "Type of network to use")

	// security
	pflag.String("seccomp", "seccomp_default.json", "Seccomp profile name")
	pflag.String("apparmor", "not-implemented", "not-implemented")

	pflag.Parse()
}