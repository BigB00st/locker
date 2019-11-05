package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// Usage: go run main.go run <cmd> <args>
func main() {
	fmt.Println("***ENTERED MAIN***")
	switch os.Args[1] {
	case "run":
		parent()
	case "child":
		child()
	default:
		panic("help")
	}
}

// Parent function, forks and execs child, which runs the requested command
func parent() {

	//fork exec self with "child" as first arg
	cmd := exec.Command("/proc/self/exe", append([]string{"child"}, os.Args[2:]...)...)

	//pipe streams
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	//namespace flags
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:   syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
		Unshareflags: syscall.CLONE_NEWNS,
	}

	must(cmd.Run())
}

// Child process, runs requested command
func child() {
	fmt.Printf("***ENTERED CHILD*** Running %v as PID: %d\n", os.Args[2:], os.Getpid())

	cmd := exec.Command(os.Args[2], os.Args[3:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	must(cmd.Run())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}