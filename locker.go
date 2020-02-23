package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gitlab.com/bigboost/locker/caps"
	"gitlab.com/bigboost/locker/cgroups"
	"gitlab.com/bigboost/locker/config"
	"gitlab.com/bigboost/locker/ipc"
	"gitlab.com/bigboost/locker/network"
	"gitlab.com/bigboost/locker/seccomp"
	"gitlab.com/bigboost/locker/utils"
)

// Usage: ./locker command args...
func main() {
	if err := config.Load(); err != nil {
		utils.PrintAndExit(err)
	}

	if os.Geteuid() != 0 {
		utils.PrintAndExit("Please run as root")
	}

	if len(pflag.Args()) < 1 {
		utils.PrintAndExit("USAGE: command args...")
	}

	if utils.IsChild() {
		if err := child(); err != nil {
			utils.PrintAndExit(err)
		}
	} else { //parent
		if err := parent(); err != nil {
			utils.PrintAndExit(err)
		}
	}
}

// Parent function, forks and execs child, which runs the requested command
func parent() error {
	/*if apparmor.Enabled() {
		if err := apparmor.InstallProfile(); err != nil {
			return err
		} else {
			defer func() {
				if err := apparmor.UnloadProfile(viper.GetString("aa-profile-path")); err != nil {
					utils.PrintAndExit(err)
				}
			}()
		}
	}*/

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
	if err := cgroups.Set(); err != nil {
		cgroups.Destruct()
		return err
	}

	//Delete new cgroups at the end
	defer func() {
		if err := cgroups.Destruct(); err != nil {
			utils.PrintAndExit(err)
		}
	}()

	if err := network.CreateConnectivity(); err != nil {
		fmt.Println(err, " - internet connectivity will be disabled")
	}

	listener, err := ipc.CreateSocket()
	if err != nil {
		return errors.Wrap(err, "couldn't create socket")
	}

	if err := cmd.Start(); err != nil {
		return errors.Wrap(err, "couldn't start child")
	}

	conn, err := ipc.SocketAccept(listener)
	if err != nil {
		return errors.Wrap(err, "couldn't create connection")
	}
	defer conn.Close()

	fmt.Println("Child PID:", cmd.Process.Pid)

	if err := cgroups.RemoveSelf(); err != nil {
		return err
	}
	if err := ipc.Send(conn, CGROUPS_ACK); err != nil {
		return err
	}

	// wait for child to request apparmor, set and send
	if err := ipc.WaitForMessage(conn, APPARMOR_REQ); err != nil {
		return err
	}
	// apparmor
	if err := ipc.Send(conn, APPARMOR_ACK); err != nil {
		return err
	}

	if err := cmd.Wait(); err != nil {
		return errors.Wrap(err, "child failed")
	}

	return nil
}

// Child process, runs requested command
func child() error {
	conn, err := ipc.ConnectToSocket()
	if err != nil {
		return errors.Wrap(err, "couldn't connect to socket")
	}
	defer conn.Close()

	// wait for cgroups
	ipc.WaitForMessage(conn, CGROUPS_ACK)

	nonFlagArgs := pflag.Args()
	fmt.Printf("Running: %v\n", nonFlagArgs[0:])

	//command to run
	cmd := exec.Command(nonFlagArgs[0], nonFlagArgs[1:]...)

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
	if err := syscall.Chdir(fsPath); err != nil {
		return errors.Wrap(err, "couldn't changedir into container")
	}
	if err := syscall.Chroot("."); err != nil {
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

	// request apparmor and wait for completion
	ipc.Send(conn, APPARMOR_REQ)
	ipc.WaitForMessage(conn, APPARMOR_ACK)

	cmd.Run()
	return nil
}
