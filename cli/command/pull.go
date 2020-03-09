package command

import (
	"os"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/web"
)

func RunPull(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker pull needs to be executed as root")
	}

	if len(args) != 1 {
		return errors.New("Usage: locker pull IMAGE")
	}
	return web.PullImage(args[0])
}
