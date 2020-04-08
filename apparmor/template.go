package apparmor

const template = `
#include <tunables/global>

$EXECUTABLE flags=(attach_disconnected) {
  #include <abstractions/base>

  $CAPS,

  file,
  network,

  deny $BASE/@{PROC}/* w,   # deny write for all files directly in /proc (not in a subdir)
  # deny write to files not in /proc/<number>/** or /proc/sys/**
  deny $BASE/@{PROC}/{[^1-9],[^1-9][^0-9],[^1-9s][^0-9y][^0-9s],[^1-9][^0-9][^0-9][^0-9]*}/** w,
  deny $BASE/@{PROC}/sys/[^k]** w,  # deny /proc/sys except /proc/sys/k* (effectively /proc/sys/kernel)
  deny $BASE/@{PROC}/sys/kernel/{?,??,[^s][^h][^m]**} w,  # deny everything except shm* in /proc/sys/kernel/
  deny $BASE/@{PROC}/sysrq-trigger rwklx,
  deny $BASE/@{PROC}/kcore rwklx,

  deny $BASE/sys/[^f]*/** wklx,
  deny $BASE/sys/f[^s]*/** wklx,
  deny $BASE/sys/fs/[^c]*/** wklx,
  deny $BASE/sys/fs/c[^g]*/** wklx,
  deny $BASE/sys/fs/cg[^r]*/** wklx,
  deny $BASE/sys/firmware/** rwklx,
  deny $BASE/sys/kernel/security/** rwklx,
}
`
