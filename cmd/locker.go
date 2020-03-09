package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func getCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "locker [OPTIONS] COMMAND [ARG...]",
		Short: "Locker is a docker-like runtime for containers",
	}

	cmdList := [](*cobra.Command){
		&cobra.Command{
			Use:   "locker run [OPTIONS] IMAGE [COMMAND] [ARG...]",
			Short: "Run a container",
		},
		&cobra.Command{
			Use:   "locker pull NAME",
			Short: "Pull an image from docker hub",
		},
		&cobra.Command{
			Use:   "locker remove NAME",
			Short: "Remove an image locally",
		},
	}

	for _, cmd := range cmdList {
		rootCmd.AddCommand(cmd)
	}
	return rootCmd
}

func Execute(cmd *cobra.Command) {
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
