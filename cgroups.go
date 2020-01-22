package main

import (
	"os"
	"io/ioutil"
	"strconv"
	"path"
	"github.com/spf13/viper"
	"github.com/spf13/pflag"
)

//cgroup function, limits recourse usage of process
func CgInit() {
	cpusAllowed := viper.GetString("cgroups.cpus-allowed")
	bytesLimit, err := ToBytes(viper.GetString("cgroups.memory-limit"))
	if err != nil {
		bytesLimit, _ = ToBytes(pflag.Lookup("memory-limit").DefValue)
	}
	swappiness := viper.GetInt("cgroups.memory-swapiness")
	
	//make cgruops
	os.Mkdir(viper.GetString("cgroups.memory-path"), 0755)
	os.Mkdir(viper.GetString("cgroups.cpuset-path"), 0755)

	//limit RAM
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), byteLimitFile), []byte(strconv.Itoa(int(bytesLimit))), 0700))

	//disable swapiness
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), swapinessFile), []byte(strconv.Itoa(swappiness)), 0700))

	//limit amount of CPUs allowes
	mems, err := ioutil.ReadFile(path.Join(viper.GetString("cgroups.cpuset-root-path"), cpusetMemFile))
	must(err)
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), cpusetMemFile), mems, 0700))
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), cpusetLimitFile), []byte(cpusAllowed), 0700))

	//assign self to memory cgroup

	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), procsFile), []byte("0"), 0700))

	//assign self to cpuset cgroup
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), procsFile), []byte("0"), 0700))

	//cleanup after container exists
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-path"), notifyOnReleaseFile), []byte("1"), 0700))
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-path"), notifyOnReleaseFile), []byte("1"), 0700))
}

func CgRemoveSelf() {
	//assign self to root memory cgroup
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.memory-root-path"), procsFile), []byte("0"), 0700))

	//assign self to root cpuset cgroup
	must(ioutil.WriteFile(path.Join(viper.GetString("cgroups.cpuset-root-path"), procsFile), []byte("0"), 0700))
}

//cgroup function, limits recourse usage of process
func CgDestruct() {
	must(os.Remove(viper.GetString("cgroups.memory-path")))
	must(os.Remove(viper.GetString("cgroups.cpuset-path")))
}
