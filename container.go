package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"github.com/spf13/pflag"
)

// Usage: ./locker command args...
func main() {
	ReadConfig()
	parseArgs()
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

	//fork exec self
	cmd := exec.Command("/proc/self/exe", pflag.Args()[0:]...)

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
	CgInit()
	defer CgDestruct()


	createNetConnectivity()

	must(cmd.Start())
	
	fmt.Println("Child PID:", cmd.Process.Pid)
	CgRemoveSelf()

	cmd.Wait()
}

// Child process, runs requested command
func child() {
	nonFlagArgs := pflag.Args()
	fmt.Printf("Running: %v\n", nonFlagArgs[0:])
	
	//command to run
	cmd := exec.Command(nonFlagArgs[0], nonFlagArgs[1:]...)

	syscallsWhitelist := readSeccompProfile(defaultSeccompProfilePath)

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(syscall.Sethostname([]byte("locker")))
	must(syscall.Chroot(fsPath))
	os.Setenv("PATH", linuxDefaultPATH)
	must(os.Chdir("/root"))

	// mount proc for pids
	must(syscall.Mount("proc", "/proc", "proc", 0, ""))
	scmpFilter := createScmpFilter(syscallsWhitelist)
	defer scmpFilter.Release()
	cmd.Run()
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
