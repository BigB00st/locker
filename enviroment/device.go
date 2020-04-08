package enviroment

import "golang.org/x/sys/unix"

// device specifies a device file
type device struct {
	path string
	mode uint32
	dev  int
}

// defaultMounts returns a list of default mounts to mount inside the container
func defaultDevices() []device {
	return []device{
		{
			path: "/dev/random",
			mode: 0666,
			dev:  int(unix.Mkdev(1, 8)),
		},
		{
			path: "/dev/urandom",
			mode: 0666,
			dev:  int(unix.Mkdev(1, 9)),
		},
		{
			path: "/dev/null",
			mode: 0666,
			dev:  int(unix.Mkdev(1, 3)),
		},
		{
			path: "/dev/zero",
			mode: 0666,
			dev:  int(unix.Mkdev(1, 5)),
		},
		{
			path: "/dev/tty",
			mode: 0666,
			dev:  int(unix.Mkdev(5, 0)),
		},
		{
			path: "/dev/console",
			mode: 0666,
			dev:  int(unix.Mkdev(136, 0)),
		},
		{
			path: "/dev/full",
			mode: 0666,
			dev:  int(unix.Mkdev(1, 7)),
		},
	}
}
