package signal

import (
	"os"
	"os/signal"
	"syscall"
)

func HandleSignals() {
	c := make(chan os.Signal)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
		syscall.SIGABRT,
		syscall.SIGSTOP,
		syscall.SIGQUIT,
		syscall.SIGCHLD,
	)
}
