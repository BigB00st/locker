package caps

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/syndtr/gocapability/capability"
	"gitlab.com/amit-yuval/locker/utils"
)

var capabilityMap = make(map[string]capability.Cap)

// init function for capabilities, maps capabilites strings to capabilites
func init() {
	last := capability.CAP_LAST_CAP
	// hack for RHEL6 which has no /proc/sys/kernel/cap_last_cap
	if last == capability.Cap(63) {
		last = capability.CAP_BLOCK_SUSPEND
	}
	for _, cap := range capability.List() {
		if cap > last {
			continue
		}
		key := "CAP_" + strings.ToUpper(cap.String())
		capabilityMap[key] = cap
	}
}

// DefaultCapabilities returns a Linux kernel default capabilities
func DefaultCapabilities() []string {
	return []string{
		"CAP_CHOWN",
		"CAP_DAC_OVERRIDE",
		"CAP_FSETID",
		"CAP_FOWNER",
		"CAP_MKNOD",
		"CAP_NET_RAW",
		"CAP_SETGID",
		"CAP_SETUID",
		"CAP_SETFCAP",
		"CAP_SETPCAP",
		"CAP_NET_BIND_SERVICE",
		"CAP_SYS_CHROOT",
		"CAP_KILL",
		"CAP_AUDIT_WRITE",
	}
}

const allCapabilities = "ALL"

// normalizeLegacyCapabilities normalizes, and validates CapAdd/CapDrop capabilities
// by upper-casing them, and adding a CAP_ prefix (if not yet present).
//
// This function also accepts the "ALL" magic-value, that's used by CapAdd/CapDrop.
func normalizeLegacyCapabilities(caps []string) ([]string, error) {
	var normalized []string

	for _, c := range caps {
		c = strings.ToUpper(c)
		if c == allCapabilities {
			normalized = append(normalized, c)
			continue
		}
		if !strings.HasPrefix(c, "CAP_") {
			c = "CAP_" + c
		}
		if _, contains := capabilityMap[c]; !contains {
			return nil, errors.Errorf("unknown capability: %q", c)
		}
		normalized = append(normalized, c)
	}
	return normalized, nil
}

func GetCapsList() ([]string, error) {
	addCaps, err := normalizeLegacyCapabilities(viper.GetStringSlice("cap-add"))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing capabilites")
	}
	dropCaps, err := normalizeLegacyCapabilities(viper.GetStringSlice("cap-drop"))
	if err != nil {
		return nil, errors.Wrap(err, "error parsing capabilites")
	}

	var caps []string
	switch {
	case utils.StringInSlice(allCapabilities, addCaps):
		// Add all capabilities except ones on dropCaps
		for k, _ := range capabilityMap {
			if !utils.StringInSlice(k, dropCaps) {
				caps = append(caps, k)
			}
		}
	case utils.StringInSlice(allCapabilities, dropCaps):
		// "Drop" all capabilities; use what's in addCaps instead
		caps = addCaps
	default:
		// First drop some capabilities
		for _, c := range DefaultCapabilities() {
			if !utils.StringInSlice(c, dropCaps) {
				caps = append(caps, c)
			}
		}
		// Then add the list of capabilities from addCaps
		caps = append(caps, addCaps...)
	}
	return caps, nil
}

// Function sets capabilites as only given list
func SetCaps(capList []string) error {
	caps, err := capability.NewPid2(0)
	if err != nil {
		return errors.Wrap(err, "couldn't initialize a new capabilities object")
	}
	for _, cap := range capList {
		caps.Set(capability.CAPS|capability.BOUNDING, capabilityMap[cap])
	}
	if err := caps.Apply(capability.CAPS | capability.BOUNDING); err != nil {
		return errors.Wrap(err, "couldn't apply capabilities")
	}
	return nil
}
