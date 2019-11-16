package main

//core related
const fsPath string= "/home/amit/containers/ubuntu"
const linuxDefaultPATH = "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

//cgroup related
const cgroupPath = "/sys/fs/cgroup/"
const cgroupMemory = "memory"
const cgroupCPU = "cpu"
const swapinessFile = "memory.swappiness"
const byteLimitFile = "memory.limit_in_bytes"
const procsFile = "cgroup.procs"
const notifyOnReleaseFile = "notify_on_release"