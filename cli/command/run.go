package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/amit-yuval/locker/apparmor"
	"gitlab.com/amit-yuval/locker/caps"
	"gitlab.com/amit-yuval/locker/cgroups"
	"gitlab.com/amit-yuval/locker/config"
	"gitlab.com/amit-yuval/locker/image"
	"gitlab.com/amit-yuval/locker/mount"
	"gitlab.com/amit-yuval/locker/network"
	"gitlab.com/amit-yuval/locker/seccomp"
	"gitlab.com/amit-yuval/locker/utils"
	"golang.org/x/sys/unix"
)

// Run runs container parent process
func Run(args []string) error {
	if os.Geteuid() != 0 {
		return errors.New("locker run needs to be executed as root")
	}
	if len(args) < 1 {
		return errors.New("Image not specified")
	}
	return parent(args)
}

// Parent function, forks and execs child, which runs the requested command
func parent(args []string) error {
	// mount image
	imageConfig, err := image.MountImage(args[0])
	if err != nil {
		return err
	}
	defer imageConfig.Cleanup()
	mergedDir := filepath.Join(imageConfig.Dir, image.Merged)

	cmdList, env, err := image.ReadConfigFile(args[0])
	if err != nil {
		return err
	}

	executablePath, err := utils.GetExecutablePath(cmdList[0], mergedDir, env)
	if err != nil {
		return err
	}

	if apparmor.Enabled() {
		profilePath, err := apparmor.Set(executablePath)
		if err != nil {
			return err
		}
		defer apparmor.UnloadProfile(profilePath)
	}

	//command to fork exec self
	cmd := exec.Command("/proc/self/exe", utils.GetChildArgs(args[0], mergedDir, cmdList)...)

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env

	//namespace flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   unix.CLONE_NEWUTS | unix.CLONE_NEWPID | unix.CLONE_NEWNS | unix.CLONE_NEWIPC | unix.CLONE_NEWCGROUP,
		Unshareflags: unix.CLONE_NEWNS | unix.CLONE_NEWUTS | unix.CLONE_NEWIPC | unix.CLONE_NEWCGROUP,
	}

	//configure cgroups
	if err := cgroups.Set(); err != nil {
		cgroups.Destruct()
		return err
	}

	//Delete new cgroups at the end
	defer func() {
		if err := cgroups.Destruct(); err != nil {
			fmt.Println(err)
		}
	}()

	netConfig, _ := network.CreateConnectivity()
	defer netConfig.Cleanup()

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "couldn't start child")
	}

	if err := cgroups.RemoveSelf(); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "child failed")
	}

	return nil
}

// Child process, runs requested command
func Child() error {
	config.Init()
	nonFlagArgs := pflag.Args()

	syscallsWhitelist, err := seccomp.ReadProfile(viper.GetString("seccomp"))
	if err != nil {
		return err
	}

	if err := syscall.Sethostname([]byte(viper.GetString("name"))); err != nil {
		return errors.Wrap(err, "couldn't set child's hostname")
	}
	if err := syscall.Chdir(nonFlagArgs[0]); err != nil {
		return errors.Wrap(err, "couldn't changedir into container")
	}
	if err := syscall.Chroot("."); err != nil {
		return errors.Wrap(err, "couldn't change root into container")
	}
	cmd := exec.Command(nonFlagArgs[1], nonFlagArgs[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := mount.MountDefaults(); err != nil {
		return err
	}

	scmpFilter, err := seccomp.CreateFilter(syscallsWhitelist)
	if err != nil {
		return err
	}
	defer scmpFilter.Release()
	if err := caps.SetCaps(viper.GetStringSlice("caps")); err != nil {
		return errors.Wrap(err, "couldn't set capabilites of child")
	}
	cmd.Run()

	return nil
}
