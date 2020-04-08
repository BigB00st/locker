package command

import (
	"os"

	"gitlab.com/amit-yuval/locker/internal/image"

	"github.com/pkg/errors"
)

// Pull pulls image
func Pull(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker pull needs to be executed as root")
	}

	if len(args) != 1 {
		return errors.New("Usage: locker pull IMAGE")
	}
	return image.PullImage(args[0])
}
