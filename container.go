package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Usage: ./locker command args...
func main() {
	err := readConfig()
	if err != nil {
		panic(err)
	}
	parseArgs()
	err = bindFlagsToConfig()
	if err != nil {
		panic(err)
	}

	if os.Geteuid() != 0 {
		fmt.Println("Please run as root")
		os.Exit(1)
	}

	if len(pflag.Args()) < 1 {
		fmt.Println("USAGE: command args...")
		os.Exit(1)
	}

	if isChild() {
		child()
	} else {
		parent()
	}
}

// Parent function, forks and execs child, which runs the requested command
func parent() {
	// drop most capabilites
	err := setCaps(setupCapabilites)
	if err != nil {
		panic(err)
	}

	if apparmorEnabled() {
		if err := InstallProfile(); err != nil {
			panic(err)
		} else {
			defer func() {
				if err := UnloadProfile(viper.GetString("aa-profile-path")); err != nil {
					panic(err)
				}
			}()
		}
	}

	//command to fork exec self
	cmd := exec.Command("/proc/self/exe", os.Args[1:]...)

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//namespace flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	//configure cgroups
	err = CgInit()
	if err != nil {
		cgerr := CgDestruct()
		if cgerr != nil {
			fmt.Println(cgerr)
		}
		panic(err)
	}

	//Delete new cgroups at the end
	defer func() {
		if err := CgDestruct(); err != nil {
			panic(err)
		}
	}()

	if err := createNetConnectivity(); err != nil {
		fmt.Println(err, " - internet connectivity will be disabled")
	}

	if err := cmd.Start(); err != nil {
		panic(errors.Wrap(err, "Couldn't start child"))
	}

	fmt.Println("Child PID:", cmd.Process.Pid)
	err = CgRemoveSelf()
	if err != nil {
		panic(err)
	}

	if err = cmd.Wait(); err != nil {
		panic(errors.Wrap(err, "Child failed"))
	}
}

// Child process, runs requested command
func child() {
	nonFlagArgs := strings.Fields(pflag.Args()[0])
	fmt.Printf("Running: %v\n", nonFlagArgs[0:])

	//command to run
	cmd := exec.Command(nonFlagArgs[0], nonFlagArgs[1:]...)

	syscallsWhitelist, err := readSeccompProfile(viper.GetString("security.seccomp"))
	if err != nil {
		panic(err)
	}

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := syscall.Sethostname([]byte(viper.GetString("name"))); err != nil {
		panic(errors.Wrap(err, "Couldn't set child's Hostname"))
	}
	if err := syscall.Chroot(fsPath); err != nil {
		panic(errors.Wrap(err, "Couldn't change root into container"))
	}
	if err := os.Setenv("PATH", viper.GetString("path")); err != nil {
		panic(errors.Wrap(err, "Couldn't set PATH environment variable"))
	}
	if err := os.Chdir("/root"); err != nil {
		panic(errors.Wrap(err, "Coldn't change directory"))
	}

	// mount proc for pids
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		panic(errors.Wrap(err, "Coudn't bind /proc"))
	}

	scmpFilter, err := createScmpFilter(syscallsWhitelist)
	if err != nil {
		panic(err)
	}
	defer scmpFilter.Release()
	if err := setCaps(containerCapabilites); err != nil {
		panic(errors.Wrap(err, "Couldn't set capabilites of child"))
	}

	_ = cmd.Run()
}
