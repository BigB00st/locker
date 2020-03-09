package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/amit-yuval/locker/cli/command"
)

func GetCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "locker [OPTIONS] COMMAND [ARG...]",
		Short: "Locker is a docker-like runtime for containers",
	}

	cmdList := [](*cobra.Command){
		&cobra.Command{
			Use:   "locker run [OPTIONS] IMAGE [COMMAND] [ARG...]",
			Short: "Run a container",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.RunRun(args)
			},
		},
		&cobra.Command{
			Use:   "locker pull NAME",
			Short: "Pull an image from docker hub",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.RunPull(args)
			},
		},
		&cobra.Command{
			Use:   "locker remove NAME",
			Short: "Remove an image locally",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.RunRemove(args)
			},
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
