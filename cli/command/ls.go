package command

import (
	"fmt"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/image"
)

// Ls lists local images
func Ls(args []string) error {
	out, err := image.ListImages()
	if err != nil {
		return errors.Wrap(err, "couldn't listing images")
	}
	fmt.Print(out)
	return nil
}
