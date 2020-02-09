package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

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
		if err := InstallProfile(); err == nil {
			defer UnloadProfile(viper.GetString("aa-profile-path"))
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
		if err != nil {
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

	must(cmd.Start())

	fmt.Println("Child PID:", cmd.Process.Pid)
	err = CgRemoveSelf()
	if err != nil {
		panic(err)
	}

	cmd.Wait()
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

	must(syscall.Sethostname([]byte(viper.GetString("name"))))
	must(syscall.Chroot(fsPath))
	os.Setenv("PATH", viper.GetString("path"))
	must(os.Chdir("/root"))

	// mount proc for pids
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))

	scmpFilter, err := createScmpFilter(syscallsWhitelist)
	if err != nil {
		panic(err)
	}
	defer scmpFilter.Release()
	setCaps(containerCapabilites)

	cmd.Run()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
