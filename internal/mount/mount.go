package mount

import (
	"os"
	"strings"

	"gitlab.com/amit-yuval/locker/pkg/io"

	"golang.org/x/sys/unix"
)

//inspired by github.com/opencontainers/runtime-spec/specs-go and docker

var flags = map[string]struct {
	clear bool
	flag  int
}{
	"defaults":      {false, 0},
	"ro":            {false, RDONLY},
	"rw":            {true, RDONLY},
	"suid":          {true, NOSUID},
	"nosuid":        {false, NOSUID},
	"dev":           {true, NODEV},
	"nodev":         {false, NODEV},
	"exec":          {true, NOEXEC},
	"noexec":        {false, NOEXEC},
	"sync":          {false, SYNCHRONOUS},
	"async":         {true, SYNCHRONOUS},
	"dirsync":       {false, DIRSYNC},
	"remount":       {false, REMOUNT},
	"mand":          {false, MANDLOCK},
	"nomand":        {true, MANDLOCK},
	"atime":         {true, NOATIME},
	"noatime":       {false, NOATIME},
	"diratime":      {true, NODIRATIME},
	"nodiratime":    {false, NODIRATIME},
	"bind":          {false, BIND},
	"rbind":         {false, RBIND},
	"unbindable":    {false, UNBINDABLE},
	"runbindable":   {false, RUNBINDABLE},
	"private":       {false, PRIVATE},
	"rprivate":      {false, RPRIVATE},
	"shared":        {false, SHARED},
	"rshared":       {false, RSHARED},
	"slave":         {false, SLAVE},
	"rslave":        {false, RSLAVE},
	"relatime":      {false, RELATIME},
	"norelatime":    {true, RELATIME},
	"strictatime":   {false, STRICTATIME},
	"nostrictatime": {true, STRICTATIME},
}

// Mount specifies a mount for a container.
type Mount struct {
	// Destination is the absolute path where the mount will be placed in the container.
	Destination string
	// Type specifies the mount kind.
	Type string
	// Source specifies the source path of the mount.
	Source string
	// Options are fstab style mount options.
	Options []string
}

// defaultMounts returns a list of default mounts to mount inside the container
func defaultMounts() []Mount {
	return []Mount{
		{
			Destination: "/proc",
			Type:        "proc",
			Source:      "proc",
			Options:     []string{"nosuid", "noexec", "nodev"},
		},
		{
			Destination: "/dev",
			Type:        "tmpfs",
			Source:      "tmpfs",
			Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
		},
		{
			Destination: "/dev/pts",
			Type:        "devpts",
			Source:      "devpts",
			Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
		},
		{
			Destination: "/sys",
			Type:        "sysfs",
			Source:      "sysfs",
			Options:     []string{"nosuid", "noexec", "nodev", "ro"},
		},
		{
			Destination: "/sys/fs/cgroup",
			Type:        "cgroup",
			Source:      "cgroup",
			Options:     []string{"ro", "nosuid", "noexec", "nodev"},
		},
		{
			Destination: "/dev/mqueue",
			Type:        "mqueue",
			Source:      "mqueue",
			Options:     []string{"nosuid", "noexec", "nodev"},
		},
		{
			Destination: "/dev/shm",
			Type:        "tmpfs",
			Source:      "shm",
			Options:     []string{"nosuid", "noexec", "nodev", "mode=1777"},
		},
	}
}

// MountDefaults mounts the default mounts
func MountDefaults() error {
	for _, v := range defaultMounts() {
		if !io.FileExists(v.Destination) {
			os.MkdirAll(v.Destination, os.ModeDir)
		}
		flag, options := parseOptions(v.Options)
		if err := unix.Mount(v.Source, v.Destination, v.Type, uintptr(flag), options); err != nil {
			continue //TODO fix cgroup error
		}
	}
	return nil
}

// parseOptions parses given options, returns mount flag and mount options string
func parseOptions(options []string) (int, string) {
	var (
		flag int
		data []string
	)

	for _, o := range options {
		// If the option does not exist in the flags table or the flag
		// is not supported on the platform,
		// then it is a data value for a specific fs type
		if f, exists := flags[o]; exists && f.flag != 0 {
			if f.clear {
				flag &= ^f.flag
			} else {
				flag |= f.flag
			}
		} else {
			data = append(data, o)
		}
	}
	return flag, strings.Join(data, ",")
}
