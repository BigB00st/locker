package main

import (
	"github.com/spf13/pflag"
)

func parseArgs() {

	// cgroup
	pflag.String("memory-limit", "1GB", "Byte must be a positive integer with measurement unit (MB, MiB, GB...)")
	pflag.Bool("memory-swappiness", false, "Allow swappiness in container")
	pflag.Parse()
}