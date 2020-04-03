package apparmor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gitlab.com/amit-yuval/locker/utils"
)

// Set sets apparmor profile if enabled, returns path of apparmor profile
func Set(executable string) (string, error) {
	apparmorPath, err := installProfile(executable)
	if err != nil {
		return "", err
	}
	return apparmorPath, nil
}

// Enabled returns true if apparmor is enabled
func Enabled() bool {
	enabled, err := utils.CmdOut("aa-enabled")
	if err != nil {
		return false
	}

	return strings.Contains(enabled, "Yes")
}

// LoadProfile runs `apparmor_parser -Kr` on a specified apparmor profile to
// replace the profile. The `-K` is necessary to make sure that apparmor_parser
// doesn't try to write to a read-only filesystem.
func loadProfile(profilePath string) error {
	if err := exec.Command("apparmor_parser", "-Kr", profilePath).Run(); err != nil {
		return errors.Wrap(err, "error loading apparmor profile")
	}
	return nil
}

// LoadProfile runs `apparmor_parser -R` on a specified apparmor profile to
// unload the profile, and deletes the file
func UnloadProfile(profilePath string) error {
	if err := exec.Command("apparmor_parser", "-R", profilePath).Run(); err != nil {
		return errors.Wrap(err, "error unloading apparmor profile")
	}
	if err := os.Remove(profilePath); err != nil {
		return errors.Wrap(err, "couldn't remove apparmor tempfile")
	}
	return nil
}

// installProfile installs default apparmor profile
func installProfile(executable string) (string, error) {
	f, err := ioutil.TempFile("", "locker")
	if err != nil {
		return "", errors.Wrap(err, "couldn't generate temp apparmor file")
	}
	defer f.Close()

	profilePath := f.Name()
	viper.Set("aa-profile-path", profilePath)

	if err := generateProfile(f, executable); err != nil {
		return "", errors.Wrap(err, "couldn't generate apparmor profile")
	}
	if err := loadProfile(profilePath); err != nil {
		return "", errors.Wrap(err, "couldn't load apparmor profile")
	}
	return f.Name(), nil
}

// generateProfile generates apparmor profile from template
func generateProfile(f *os.File, executable string) error {
	profile := template
	profile = strings.Replace(profile, "$EXECUTABLE", executable, 1)
	profile = strings.Replace(profile, "$CAPS", getCaps(), 1)

	if _, err := f.Write([]byte(profile)); err != nil {
		return errors.Wrap(err, "error writing to apparmor profile")
	}
	return nil
}

// getCaps gets capabilities in format of apparmor profile
func getCaps() string {
	return "capability " + strings.ToLower(strings.ReplaceAll(strings.Join(viper.GetStringSlice("caps"), ",capability "), "CAP_", ""))
}
