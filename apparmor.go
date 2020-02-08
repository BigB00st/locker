package main

import (
	"strings"
)

func apparmorEnabled() bool {
	enabled, err := cmdOut("aa-enabled")
	if err != nil {
		return false
	}

	return strings.Contains(enabled, "Yes")
}
