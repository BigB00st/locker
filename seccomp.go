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

func createScmpFilter(syscalls []string) *libseccomp.ScmpFilter {
	// blacklist everything (EPERM - Permission not permitted)
	scmpFilter, err := libseccomp.NewFilter(libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	must(err)
	
	// whitelist given syscalls 
    for _, syscall := range syscalls {
        syscallID, err := libseccomp.GetSyscallFromName(syscall)
		if err == nil {
			must(scmpFilter.AddRule(syscallID, libseccomp.ActTrace))
		}
	}
	
	must(scmpFilter.Load())
	return scmpFilter
}

func resetScmpFilter(scmpFilter *libseccomp.ScmpFilter) {
	must(scmpFilter.Reset(libseccomp.ActTrace))
	must(scmpFilter.Load())
}