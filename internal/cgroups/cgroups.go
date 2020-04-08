package cgroups

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"gitlab.com/amit-yuval/locker/internal/utils"

	"code.cloudfoundry.org/bytefmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"golang.org/x/sys/unix"
)

const (
	basePath          = "/sys/fs/cgroup/"
	memoryPath        = "memory"
	pidsPath          = "pids"
	cpuSetPath        = "cpuset"
	swapinessFile     = "memory.swappiness"
	byteLimitFile     = "memory.limit_in_bytes"
	kmemByteLimitFile = "memory.kmem.limit_in_bytes"
	tcpByteLimitFile  = "memory.kmem.tcp.limit_in_bytes"
	cpusetLimitFile   = "cpuset.cpus"
	cpusetMemFile     = "cpuset.mems"
	procsFile         = "cgroup.procs"
	pidsFile          = "pids.max"
	minMemory         = 5000000
	minPids           = 10
)

// init sets directory names for cgroups
func init() {
	viper.Set("cgroup-name", "locker"+strconv.Itoa(os.Getpid()))
	viper.Set("cpuset-path", path.Join(basePath, cpuSetPath, viper.GetString("cgroup-name")))
	viper.Set("cpuset-root-path", path.Join(basePath, cpuSetPath))
	viper.Set("memory-path", path.Join(basePath, memoryPath, viper.GetString("cgroup-name")))
	viper.Set("memory-root-path", path.Join(basePath, memoryPath))
	viper.Set("pids-path", path.Join(basePath, pidsPath, viper.GetString("cgroup-name")))
	viper.Set("pids-root-path", path.Join(basePath, pidsPath))
}

// Set limits recourse usage of process by setting cgroup rules
func Set() error {
	cpusAllowed := viper.GetString("cpus-allowed")
	swappiness := strconv.Itoa(viper.GetInt("memory-swappiness"))
	maxPids := strconv.Itoa(utils.Max(viper.GetInt("max-pids"), minPids))
	bytesLimit, err := bytefmt.ToBytes(viper.GetString("memory-limit"))
	if err != nil {
		return errors.Wrap(err, "couldn't parse memory-limit")
	}
	memoryLimit := strconv.Itoa(utils.Max(int(bytesLimit), minMemory))

	// make cgroup directories
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := os.Mkdir(fileName, os.ModeDir); err != nil {
			return errors.Wrapf(err, "couldn't make cgroup directory %v", fileName)
		}
	}

	// set swappiness
	if err := ioutil.WriteFile(path.Join(viper.GetString("memory-path"), swapinessFile), []byte(swappiness), 0700); err != nil {
		return errors.Wrapf(err, "couldn't write %q to the the memory's swappiness file", swappiness)
	}

	// limit amount of CPUs allowes
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

	// limit memory
	for _, fileName := range []string{byteLimitFile, kmemByteLimitFile, tcpByteLimitFile} {
		if err := ioutil.WriteFile(path.Join(viper.GetString("memory-path"), fileName), []byte(memoryLimit), 0700); err != nil {
			return errors.Wrapf(err, "couldn't write %v to %v", []byte(memoryLimit), fileName)
		}
	}

	// assign self to cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := ioutil.WriteFile(path.Join(fileName, procsFile), []byte("0"), 0700); err != nil {
			return errors.Wrapf(err, "couldn't assign self to new %v cgroup", fileName)
		}
	}

	return nil
}

// RemoveSelf moves current process to root cgroups
func RemoveSelf() error {
	//assign self to root memory cgroup
	if err := ioutil.WriteFile(path.Join(viper.GetString("memory-root-path"), procsFile), []byte("0"), 0700); err != nil {
		return errors.Wrap(err, "couldn't assign parent process to root memory cgroup")
	}

	// assign self to root cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-root-path"), viper.GetString("cpuset-root-path"), viper.GetString("pids-root-path")} {
		if err := ioutil.WriteFile(path.Join(fileName, procsFile), []byte("0"), 0700); err != nil {
			return errors.Wrapf(err, "couldn't assign self to root %v cgroup", fileName)
		}
	}

	return nil
}

// Destruct cleans cgroups
func Destruct() error {
	// assign self to cgroups by writing "0" to procs file
	for _, fileName := range []string{viper.GetString("memory-path"), viper.GetString("cpuset-path"), viper.GetString("pids-path")} {
		if err := unix.Rmdir(fileName); err != nil {
			return errors.Wrapf(err, "couldn't remove %v", fileName)
		}
	}

	return nil
}
