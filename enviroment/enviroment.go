package enviroment

import (
	"os/exec"
	"path/filepath"

	"github.com/pkg/errors"
	"gitlab.com/amit-yuval/locker/utils"
	"golang.org/x/sys/unix"
)

func CopyFiles(baseDir string) error {
	files := []string{"/etc/resolv.conf"}
	for _, file := range files {
		src, dst := file, filepath.Join(baseDir, file)
		utils.MkdirIfNotExist(filepath.Dir(dst))
		if _, err := utils.Copy(src, dst); err != nil {
			return errors.Wrapf(err, "couldn't copy %v", src)
		}
	}
	return nil
}

// Setup calls inner setup function for the enviroment
func Setup() {
	createDevices()
	configLinker()
}

// createDevices creates default devies
func createDevices() {
	for _, device := range defaultDevices() {
		unix.Mknod(device.path, unix.S_IFCHR|device.mode, device.dev)
		unix.Chmod(device.path, device.mode)
	}
}

// configLinker invokes "ldconfig" command which configures dynamic
// linker run-time bindings
func configLinker() {
	exec.Command("ldconfig").Run()
}
