package main

import (
	"github.com/pkg/errors"
	"github.com/syndtr/gocapability/capability"
)

var setupCapabilites []capability.Cap = []capability.Cap{
	//capability.CAP_AUDIT_CONTROL,
	//capability.CAP_AUDIT_READ,
	//capability.CAP_AUDIT_WRITE,
	//capability.CAP_BLOCK_SUSPEND,
	//capability.CAP_CHOWN,
	capability.CAP_DAC_OVERRIDE, // needed to bypass rwx permissions
	//capability.CAP_DAC_READ_SEARCH,
	//capability.CAP_FOWNER,
	//capability.CAP_FSETID,
	//capability.CAP_IPC_LOCK,
	//capability.CAP_KILL,
	//capability.CAP_LEASE,
	//capability.CAP_LINUX_IMMUTABLE,
	capability.CAP_MAC_ADMIN,    //needed for MAC
	capability.CAP_MAC_OVERRIDE, //needed for MAC
	//capability.CAP_MKNOD,
	capability.CAP_NET_ADMIN,        //needed for network
	capability.CAP_NET_BIND_SERVICE, //needed for port binding (<1024)
	//capability.CAP_NET_BROADCAST,
	capability.CAP_NET_RAW, //needed for network
	//capability.CAP_SETGID,
	//capability.CAP_SETFCAP,
	//capability.CAP_SETPCAP,
	//capability.CAP_SETUID,
	capability.CAP_SYS_ADMIN, //mount, sethostname, employ clone flags, setns, set seccomp filter, modify cgroups
	//capability.CAP_SYS_BOOT,
	capability.CAP_SYS_CHROOT, //needed for chrootns
	//capability.CAP_SYS_MODULE,
	capability.CAP_SYS_NICE, //needed for writing cpuset cgroup
	//capability.CAP_SYS_PACCT,
	//capability.CAP_SYS_PTRACE,
	//capability.CAP_SYS_RAWIO,
	//capability.CAP_SYS_RESOURCE,
	//capability.CAP_SYS_TIME,
	//capability.CAP_SYS_TTY_CONFIG,
	//capability.CAP_SYSLOG,
	//capability.CAP_WAKE_ALARM,
}

var containerCapabilites []capability.Cap = []capability.Cap{
	capability.CAP_NET_ADMIN, //needed for network
	capability.CAP_NET_RAW,   //needed for network
}

// Function sets capabilites as only given list
func setCaps(capList []capability.Cap) error {
	caps, err := capability.NewPid2(0)
	if err != nil {
		return errors.Wrap(err, "couldn't initialize a new capabilities object")
	}
	for _, cur := range capList {
		caps.Set(capability.CAPS|capability.BOUNDING, cur)
	}
	err = caps.Apply(capability.CAPS | capability.BOUNDING)
	if err != nil {
		return errors.Wrap(err, "couldn't apply capabilities")
	}
	return nil
}
