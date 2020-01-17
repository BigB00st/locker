package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Usage: ./locker command args...
func main() {
	if len(os.Args) < 2 {
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
	config := NewConfig()
	CgInit(config)
	defer CgDestruct(config)


	createNetConnectivity()

	must(cmd.Start())
	
	fmt.Println("Child PID:", cmd.Process.Pid)
	CgRemoveSelf(config)

	cmd.Wait()
}

// Child process, runs requested command
func child() {
	fmt.Printf("Running: %v\n", os.Args[1:])
	
	//command to run
	cmd := exec.Command(os.Args[1], os.Args[2:]...)

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
