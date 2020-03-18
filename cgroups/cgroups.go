package cgroups

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"syscall"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

const (
	BasePath        = "/sys/fs/cgroup/"
	MemoryPath      = "memory"
	PidsPath        = "pids"
	CPUSetPath      = "cpuset" //Need to decide if we also want cpu control
	swapinessFile   = "memory.swappiness"
	byteLimitFile   = "memory.limit_in_bytes"
	selfMinMemory   = 5000000 //required for loading the parent
	cpusetLimitFile = "cpuset.cpus"
	cpusetMemFile   = "cpuset.mems"
	procsFile       = "cgroup.procs"
	pidsFile        = "pids.max"
)

func init() {
	viper.Set("cgroups.name", "locker"+strconv.Itoa(os.Getpid()))
	viper.Set("cgroups.cpuset-path", path.Join(BasePath, CPUSetPath, viper.GetString("cgroups.name")))
	viper.Set("cgroups.cpuset-root-path", path.Join(BasePath, CPUSetPath))
	viper.Set("cgroups.memory-path", path.Join(BasePath, MemoryPath, viper.GetString("cgroups.name")))
	viper.Set("cgroups.memory-root-path", path.Join(BasePath, MemoryPath))
	viper.Set("cgroups.pids-path", path.Join(BasePath, PidsPath, viper.GetString("cgroups.name")))
	viper.Set("cgroups.pids-root-path", path.Join(BasePath, PidsPath))
}

//cgroup function, limits recourse usage of process
func Set() error {
	cpusAllowed := viper.GetString("cgroups.cpus-allowed")
	bytesLimit, err := bytefmt.ToBytes(viper.GetString("cgroups.memory-limit"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse memory-limit")
	}
	swappiness := 0
	if viper.GetBool("cgroups.memory-swappiness") {
		swappiness = 1
	}
	maxPids := viper.GetString("cgroups.max-pids")

	//make memory directory
	if err := os.Mkdir(viper.GetString("cgroups.memory-path"), os.ModeDir); err != nil {
		return errors.Wrapf(err, "couldn't make memory's cgroup at %q", viper.GetString("cgroups.memory-path"))
	}
	//make cpuset directory
	if err := os.Mkdir(viper.GetString("cgroups.cpuset-path"), os.ModeDir); err != nil {
		return errors.Wrapf(err, "couldn't make cpuset's cgroup at %q", viper.GetString("cgroups.cpuset-path"))
	}
	//make pids directory
	if err := os.Mkdir(viper.GetString("cgroups.pids-path"), os.ModeDir); err != nil {
		return errors.Wrapf(err, "couldn't make pids' cgroup at %q", viper.GetString("cgroups.pids-path"))
	}

	//limit RAM and allow more for parent process
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), byteLimitFile), []byte(strconv.Itoa(int(bytesLimit+selfMinMemory))), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the the memory's byte-limit file", []byte(strconv.Itoa(int(bytesLimit+selfMinMemory))))
	}

	//disable swapiness
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), swapinessFile), []byte(strconv.Itoa(swappiness)), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the the memory's swappiness file", swappiness)
	}

	//limit amount of CPUs allowes
	mems, err := ioutil.ReadFile(path.Join(viper.GetString("cgroups.cpuset-root-path"), cpusetMemFile))
	if err != nil {
		return errors.Wrapf(err, "couldn't read cpuset's default memory at %q", cpusetMemFile)
	}

	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), cpusetMemFile), mems, 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the cpuset's memory file", mems)
	}

	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), cpusetLimitFile), []byte(cpusAllowed), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the cpuset's cpus file", cpusAllowed)
	}

	// limit amount of PIDS
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.pids-path"), pidsFile), []byte(maxPids), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %v to %v", maxPids, pidsFile)
	}

	//assign self to memory cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign self to new memory cgroup")
	}
	//assign self to cpuset cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign self to new cpuset cgroup")
	}
	//assign self to pids cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.pids-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign self to new pids cgroup")
	}
	return nil
}

func RemoveSelf() error {
	//assign self to root memory cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-root-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign parent process to root memory cgroup")
	}

	//assign self to root cpuset cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-root-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign parent process to root cpuset cgroup")
	}

	//assign self to root pids cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.pids-root-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign parent process to root pids cgroup")
	}

	bytesLimit, err := bytefmt.ToBytes(viper.GetString("cgroups.memory-limit"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse memory-limit")
	}
	//relimit RAM
	if err := ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), byteLimitFile), []byte(strconv.Itoa(int(bytesLimit))), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the the memory's byte-limit file", []byte(strconv.Itoa(int(bytesLimit))))
	}
	return nil
}

//cgroup function, limits recourse usage of process
func Destruct() error {
	if err := syscall.Rmdir(viper.GetString("cgroups.memory-path")); err != nil {
		return errors.Wrap(err, "couldn't remove memory cgroup directory")
	}
	if err := syscall.Rmdir(viper.GetString("cgroups.cpuset-path")); err != nil {
		return errors.Wrap(err, "couldn't remove cpuset cgroup directory")
	}
	return nil
}
