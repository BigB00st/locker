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
	BasePath          = "/sys/fs/cgroup/"
	MemoryPath        = "memory"
	PidsPath          = "pids"
	CPUSetPath        = "cpuset"
	swapinessFile     = "memory.swappiness"
	byteLimitFile     = "memory.limit_in_bytes"
	kmemByteLimitFile = "memory.kmem.limit_in_bytes"
	tcpByteLimitFile  = "memory.kmem.tcp.limit_in_bytes"
	cpusetLimitFile   = "cpuset.cpus"
	cpusetMemFile     = "cpuset.mems"
	procsFile         = "cgroup.procs"
	pidsFile          = "pids.max"
)

func init() {
	viper.Set("cgroup-name", "locker"+strconv.Itoa(os.Getpid()))
	viper.Set("cpuset-path", path.Join(BasePath, CPUSetPath, viper.GetString("cgroup-name")))
	viper.Set("cpuset-root-path", path.Join(BasePath, CPUSetPath))
	viper.Set("memory-path", path.Join(BasePath, MemoryPath, viper.GetString("cgroup-name")))
	viper.Set("memory-root-path", path.Join(BasePath, MemoryPath))
	viper.Set("pids-path", path.Join(BasePath, PidsPath, viper.GetString("cgroup-name")))
	viper.Set("pids-root-path", path.Join(BasePath, PidsPath))
}

//cgroup function, limits recourse usage of process
func Set() error {
	cpusAllowed := viper.GetString("cpus-allowed")
	swappiness := viper.GetString("memory-swappiness")
	maxPids := viper.GetString("max-pids")
	_, err := bytefmt.ToBytes(viper.GetString("memory-limit"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse memory-limit")
	}

	// make cgroup directories
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := os.Mkdir(fileName, os.ModeDir); err != nil {
			return errors.Wrapf(err, "couldn't make cgroup directory %v", fileName)
		}
	}

	//set swapiness
	if err := ioutil.WriteFile(path.Join(viper.GetString("memory-path"), swapinessFile), []byte(swappiness), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the the memory's swappiness file", swappiness)
	}

	//limit amount of CPUs allowes
	mems, err := ioutil.ReadFile(path.Join(viper.GetString("cpuset-root-path"), cpusetMemFile))
	if err != nil {
		return errors.Wrapf(err, "couldn't read cpuset's default memory at %q", cpusetMemFile)
	}

	if err := ioutil.WriteFile(path.Join(viper.GetString("cpuset-path"), cpusetMemFile), mems, 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the cpuset's memory file", mems)
	}

	if err := ioutil.WriteFile(path.Join(viper.GetString("cpuset-path"), cpusetLimitFile), []byte(cpusAllowed), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the cpuset's cpus file", cpusAllowed)
	}

	// limit amount of PIDS
	if err := ioutil.WriteFile(path.Join(viper.GetString("pids-path"), pidsFile), []byte(maxPids), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %v to %v", maxPids, pidsFile)
	}

	// assign self to cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := ioutil.WriteFile(path.Join(fileName, procsFile), []byte("0"), 0700); err != nil {
			return errors.Wrapf(err, "couldn't assign self to new %v cgroup", fileName)
		}
	}

	return nil
}

func RemoveSelf() error {
	bytesLimit, _ := bytefmt.ToBytes(viper.GetString("memory-limit"))

	//assign self to root memory cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("memory-root-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign parent process to root memory cgroup")
	}

	//limit memory
	for _, fileName := range []string{byteLimitFile, kmemByteLimitFile, tcpByteLimitFile} {
		if err := ioutil.WriteFile(path.Join(viper.GetString("memory-path"), fileName), []byte(strconv.Itoa(int(bytesLimit))), 0700); err != nil {
			return errors.Wrapf(err, "couldn't write %v to %v", []byte(strconv.Itoa(int(bytesLimit))), fileName)
		}
	}

	// assign self to root cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-root-path"), viper.GetString("cpuset-root-path"), viper.GetString("pids-root-path")} {
		if err := ioutil.WriteFile(path.Join(fileName, procsFile), []byte("0"), 0700); err != nil {
			return errors.Wrapf(err, "couldn't assign self to root %v cgroup", fileName)
		}
	}

	return nil
}

//cgroup function
func Destruct() error {
	// assign self to cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := syscall.Rmdir(fileName); err != nil {
			return errors.Wrapf(err, "couldn't remove %v", fileName)
		}
	}

	return nil
}
