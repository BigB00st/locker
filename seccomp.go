package main

import (
	"io/ioutil"
	"encoding/json"
	"syscall"
	libseccomp "github.com/seccomp/libseccomp-golang"
)

func readSeccompProfile(path string) []string {
	data, err := ioutil.ReadFile(path)
	must(err)

	var result map[string][]string

	// load json into AllowedSyscalls struct
	must(json.Unmarshal(data, &result))

	return result["syscalls"]
}

func seccompWhitelist(syscalls []string) {
	// blacklist everything (EPERM - Permission not permitted)
	filter, err := libseccomp.NewFilter(libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	must(err)
	
	// whitelist given syscalls 
    for _, element := range syscalls {
        syscallID, _ := libseccomp.GetSyscallFromName(element)

        filter.AddRule(syscallID, libseccomp.ActAllow)
	}
	// load seccomp filter
    filter.Load()
}
