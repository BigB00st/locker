package config

import (
	"github.com/spf13/pflag"
)

// parseArgs parses arguments
func parseArgs() {
	// generic
	pflag.String("name", "locker", "Name of container (used in hostname and more)")

	// cgroups
	pflag.String("memory-limit", "1GB", "RAM limit of container in bytes")
	pflag.Int("memory-swappiness", 30, "Memory swappiness in container")
	pflag.String("cpus-allowed", "0", "Number of cpu cores to use in container")
	pflag.Int("max-pids", 100, "Maximum number of pids available in the container")

	// security
	pflag.String("seccomp", "/etc/locker/seccomp_default.json", "Seccomp profile path")

	// capabilites
	pflag.StringSlice("cap-add", nil, "Add linux capabilites")
	pflag.StringSlice("cap-drop", nil, "Drop linux capabilites")

	pflag.Parse()
}
