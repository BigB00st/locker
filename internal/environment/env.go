package environment

import (
	"fmt"

	"github.com/spf13/viper"
)

// AppendEnv appends necessary values to env
func AppendEnv(envList []string) []string {
	term := "TERM=xterm"
	hostname := fmt.Sprintf("HOSTNAME=%v", viper.GetString("name"))
	return append(envList, term, hostname)
}
