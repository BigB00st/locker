package command

import (
	"os"

	"gitlab.com/amit-yuval/locker/internal/image"

	"github.com/pkg/errors"
)

// Remove removes image
func Remove(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker remove needs to be executed as root")
	}

	if len(args) != 1 {
		return errors.New("Usage: locker remove IMAGE")
	}
	return image.RemoveImage(args[0])
}
