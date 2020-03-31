package command

import (
	"os"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/image"
)

func Pull(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker pull needs to be executed as root")
	}

	if len(args) != 1 {
		return errors.New("Usage: locker pull IMAGE")
	}
	return image.PullImage(args[0])
}
