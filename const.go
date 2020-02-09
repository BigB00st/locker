package main

//core related
const fsPath = "/home/amit/containers/ubuntu"
const linuxDefaultPATH = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
const configFile = "config"

//cgroup related
const cgroupPath = "/sys/fs/cgroup/"
const cgroupMemory = "memory"
const cgroupCPUSet = "cpuset" //Need to decide if we also want cpu control
const swapinessFile = "memory.swappiness"
const byteLimitFile = "memory.limit_in_bytes"
const selfMinMemory = 5000000 //required for loading the parent
const cpusetLimitFile = "cpuset.cpus"
const cpusetMemFile = "cpuset.mems"
const procsFile = "cgroup.procs"

// network related
const subnetMaskBytes = 4
const subnetLogicOne = 255
const bitsInByte = 8
const ipRouteDefaultIndex = 0
const ipRouteNameIndex = 4
const netnsDirectory = "/var/run/netns/"
