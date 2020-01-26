package main

import (
	"github.com/syndtr/gocapability/capability"
)

var capabilitesBlacklist []capability.Cap = []capability.Cap{
	capability.CAP_AUDIT_CONTROL,
	capability.CAP_AUDIT_READ,
	capability.CAP_AUDIT_WRITE,
	capability.CAP_BLOCK_SUSPEND,
	capability.CAP_DAC_READ_SEARCH,
	capability.CAP_FSETID,
	capability.CAP_IPC_LOCK,
	capability.CAP_MAC_ADMIN,
	capability.CAP_MAC_OVERRIDE,
	capability.CAP_MKNOD,
	capability.CAP_SETFCAP,
	capability.CAP_SYSLOG,
	capability.CAP_SYS_ADMIN,
	capability.CAP_SYS_BOOT,
	capability.CAP_SYS_MODULE,
	capability.CAP_SYS_NICE,
	capability.CAP_SYS_RAWIO,
	capability.CAP_SYS_RESOURCE,
	capability.CAP_SYS_TIME,
	capability.CAP_WAKE_ALARM,
}

func dropCaps() {
	caps, err := capability.NewPid2(0)
	must(err)
	err = caps.Load()
	must(err)
	for _, cur := range capabilitesBlacklist {
		caps.Unset(capability.CAPS | capability.BOUNDING, cur)
	}
	caps.Apply(capability.CAPS | capability.BOUNDING)
}
