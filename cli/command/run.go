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

func RunRun(args []string) error {
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
	err := image.MountImage(args[0])
	defer image.Cleanup(args[0])
	if err != nil {
		return err
	}

	cmdList, env, err := image.ReadConfigFile(args[0])
	if err != nil {
		return err
	}

	executablePath, err := utils.GetExecutablePath(cmdList[0], filepath.Join(image.ImagesDir, args[0], image.Merged), env)
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

	//command to fork exec selfcmdList
	cmd := exec.Command("/proc/self/exe", utils.GetChildArgs(cmdList)...)

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

	if err := network.CreateConnectivity(); err != nil {
		fmt.Println(err, " - internet connectivity will be disabled")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "couldn't start child")
	}

	fmt.Println("Child PID:", cmd.Process.Pid)

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
	fmt.Println("Running:", nonFlagArgs[1:])

	syscallsWhitelist, err := seccomp.ReadProfile(viper.GetString("seccomp"))
	if err != nil {
		return err
	}
	if err := syscall.Sethostname([]byte(viper.GetString("name"))); err != nil {
		return errors.Wrap(err, "couldn't set child's hostname")
	}
	if err := syscall.Chdir(filepath.Join(image.ImagesDir, nonFlagArgs[0], image.Merged)); err != nil {
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
