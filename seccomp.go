package main

import (
	"encoding/json"
	"io/ioutil"
	"syscall"

	"github.com/pkg/errors"
	libseccomp "github.com/seccomp/libseccomp-golang"
)

func readSeccompProfile(path string) ([]string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't read seccomp profile")
	}

	var result map[string][]string

	// load json into AllowedSyscalls struct
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse seccomp profile")
	}

	return result["syscalls"], nil
}

func createScmpFilter(syscalls []string) (*libseccomp.ScmpFilter, error) {
	// blacklist everything (EPERM - Permission not permitted)
	scmpFilter, err := libseccomp.NewFilter(libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	if err != nil {
		return nil, errors.Wrap(err, "couldn't create new seccomp filter")
	}

	// whitelist given syscalls
	for _, syscall := range syscalls {
		syscallID, err := libseccomp.GetSyscallFromName(syscall)
		if err == nil {
			err = scmpFilter.AddRule(syscallID, libseccomp.ActAllow)
			if err != nil {
				return nil, errors.Wrapf(err, "couldn't allow %q syscall in seccomp profile", syscall)
			}
		}
	}

	err = scmpFilter.Load()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load seccomp profile")
	}
	return scmpFilter, nil
}
