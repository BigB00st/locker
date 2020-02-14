package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

// Function returns true if apparmor is enabled
func apparmorEnabled() bool {
	enabled, err := cmdOut("aa-enabled")
	if err != nil {
		return false
	}

	return strings.Contains(enabled, "Yes")
}

// LoadProfile runs `apparmor_parser -Kr` on a specified apparmor profile to
// replace the profile. The `-K` is necessary to make sure that apparmor_parser
// doesn't try to write to a read-only filesystem.
func LoadProfile(profilePath string) error {
	if err := exec.Command("apparmor_parser", "-Kr", profilePath).Run(); err != nil {
		return errors.Wrap(err, "error loading apparmor profile")
	}
	return nil
}

// LoadProfile runs `apparmor_parser -R` on a specified apparmor profile to
// unload the profile.
func UnloadProfile(profilePath string) error {
	if err := exec.Command("apparmor_parser", "-R", profilePath).Run(); err != nil {
		return errors.Wrap(err, "error unloading apparmor profile")
	}
	if err := os.Remove(profilePath); err != nil {
		return errors.Wrap(err, "couldn't remove apparmor tempfile")
	}
	return nil
}

// Function installs default apparmor profile
func InstallProfile() error {
	f, err := ioutil.TempFile("", viper.GetString("security.aa-profile-name"))
	if err != nil {
		return errors.Wrap(err, "couldn't generate temp apparmor file")
	}
	defer f.Close()

	profilePath := f.Name()
	viper.Set("aa-profile-path", profilePath)

	if err := GenerateProfile(f); err != nil {
		return errors.Wrap(err, "couldn't generate apparmor profile")
	}
	if err := LoadProfile(profilePath); err != nil {
		return errors.Wrap(err, "couldn't load apparmor profile")
	}
	return nil
}

func GenerateProfile(f *os.File) error {
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	profileBytes, err := ioutil.ReadFile(viper.GetString("security.aa-template"))
	if err != nil {
		return err
	}

	profile := string(profileBytes)
	profile = strings.Replace(profile, "$EXECUTABLE", ex, 1)
	profile = strings.Replace(profile, "$TEMP-FILE", f.Name(), 1)
	profile = strings.Replace(profile, "$COMMAND", os.Args[1], 1)

	_, err = f.Write([]byte(profile))
	return err
}
