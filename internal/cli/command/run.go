package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"gitlab.com/amit-yuval/locker/internal/apparmor"
	"gitlab.com/amit-yuval/locker/internal/caps"
	"gitlab.com/amit-yuval/locker/internal/cgroups"
	"gitlab.com/amit-yuval/locker/internal/config"
	"gitlab.com/amit-yuval/locker/internal/enviroment"
	"gitlab.com/amit-yuval/locker/internal/image"
	"gitlab.com/amit-yuval/locker/internal/mount"
	"gitlab.com/amit-yuval/locker/internal/network"
	"gitlab.com/amit-yuval/locker/internal/seccomp"
	"gitlab.com/amit-yuval/locker/internal/utils"

	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	env = enviroment.AppendEnv(env)

	executablePath, err := utils.GetExecutablePath(cmdList[0], mergedDir, env)
	if err != nil {
		return err
	}

	if apparmor.Enabled() {
		profilePath, err := apparmor.Set(mergedDir, executablePath)
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
	cmd.SysProcAttr = &unix.SysProcAttr{
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
	baseDir, executable := nonFlagArgs[0], nonFlagArgs[1]

	if err := enviroment.CopyFiles(baseDir); err != nil {
		return err
	}

	unixsWhitelist, err := seccomp.ReadProfile(viper.GetString("seccomp"))
	if err != nil {
		return err
	}

	if err := unix.Sethostname([]byte(viper.GetString("name"))); err != nil {
		return errors.Wrap(err, "couldn't set child's hostname")
	}
	if err := unix.Chdir(baseDir); err != nil {
		return errors.Wrap(err, "couldn't changedir into container")
	}
	if err := unix.Chroot("."); err != nil {
		return errors.Wrap(err, "couldn't change root into container")
	}
	cmd := exec.Command(executable, nonFlagArgs[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := mount.MountDefaults(); err != nil {
		return err
	}

	enviroment.Setup()

	scmpFilter, err := seccomp.CreateFilter(unixsWhitelist)
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
