package command

import (
	"os"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/web"
)

func RunRemove(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker remove needs to be executed as root")
	}

	if len(args) != 1 {
		return errors.New("Usage: locker remove IMAGE")
	}
	return web.RemoveImage(args[0])
}
