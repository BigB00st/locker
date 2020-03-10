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
	"gitlab.com/amit-yuval/locker/caps"
	"gitlab.com/amit-yuval/locker/cgroups"
	"gitlab.com/amit-yuval/locker/config"
	"gitlab.com/amit-yuval/locker/image"
	"gitlab.com/amit-yuval/locker/network"
	"gitlab.com/amit-yuval/locker/seccomp"
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
	/*if apparmor.Enabled() {
		if err := apparmor.InstallProfile(); err != nil {
			return err
		} else {
			defer func() {
				if err := apparmor.UnloadProfile(viper.GetString("aa-profile-path")); err != nil {
					fmt.Println(err)
				}
			}()
		}
	}*/
	//command to fork exec self
	cmd := exec.Command("/proc/self/exe", args...)

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//namespace flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	err := image.MountImage(args[0])
	if err != nil {
		return err
	}
	defer image.Cleanup(args[0])

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
	fmt.Println("Running:", nonFlagArgs[0:])

	//command to run
	cmd := exec.Command("/bin/bash")

	syscallsWhitelist, err := seccomp.ReadProfile(viper.GetString("security.seccomp"))
	if err != nil {
		return err
	}

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := syscall.Sethostname([]byte(viper.GetString("name"))); err != nil {
		return errors.Wrap(err, "couldn't set child's hostname")
	}
	if err := syscall.Chdir("/home/amit/containers/ubuntu"); err != nil {
		return errors.Wrap(err, "couldn't changedir into container")
	}
	if err := syscall.Chroot(filepath.Join(image.ImagesDir, nonFlagArgs[0], image.Merged)); err != nil {
		return errors.Wrap(err, "couldn't change root into container")
	}
	if err := os.Setenv("PATH", viper.GetString("path")); err != nil {
		return errors.Wrap(err, "couldn't set PATH environment variable")
	}
	if err := os.Chdir("/root"); err != nil {
		return errors.Wrap(err, "couldn't change directory to /root in container")
	}

	// mount proc for pids
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return errors.Wrap(err, "couldn't mount /proc")
	}

	scmpFilter, err := seccomp.CreateFilter(syscallsWhitelist)
	if err != nil {
		return err
	}
	defer scmpFilter.Release()
	if err := caps.SetCaps(viper.GetStringSlice("security.caps")); err != nil {
		return errors.Wrap(err, "couldn't set capabilites of child")
	}
	cmd.Run()

	return nil
}
