package signal

import (
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

func HandleSignals() {
	c := make(chan os.Signal)
	signal.Notify(c,
		unix.SIGINT,
		unix.SIGTERM,
		unix.SIGHUP,
		unix.SIGABRT,
		unix.SIGSTOP,
		unix.SIGQUIT,
		unix.SIGCHLD,
	)
}
