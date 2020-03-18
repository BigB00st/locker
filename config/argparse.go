package config

import (
	"github.com/spf13/pflag"
)

func parseArgs() {
	// generic
	pflag.String("name", "locker", "Name of container (used in hostname and more)")

	// cgroups
	pflag.String("memory-limit", "1GB", "RAM limit of container in bytes")
	pflag.Bool("memory-swappiness", false, "Allow swappiness in container (boolean)")
	pflag.String("cpus-allowed", "0", "Number of cpu cores to use in container")
	pflag.String("max-pids", "100", "Maximum number of pids available in the container")

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
