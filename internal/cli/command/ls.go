package command

import (
	"fmt"
	"os"

	"gitlab.com/amit-yuval/locker/internal/image"

	"github.com/pkg/errors"
)

// Ls lists local images
func Ls(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker ls needs to be executed as root")
	}

	out, err := image.ListImages()
	if err != nil {
		return errors.Wrap(err, "couldn't list images")
	}
	fmt.Print(out)
	return nil
}
