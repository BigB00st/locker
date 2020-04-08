package cmd

import (
	"github.com/spf13/cobra"
	"gitlab.com/amit-yuval/locker/cli/command"
	"gitlab.com/amit-yuval/locker/config"
)

// GetCmd returns cobra cmd of all commands
func GetCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "locker [OPTIONS] COMMAND [ARG...]",
		Short:        "Locker is a docker-like runtime for containers",
		SilenceUsage: true,
	}

	cmdList := [](*cobra.Command){
		&cobra.Command{
			Use:   "run [OPTIONS] IMAGE [COMMAND] [ARG...]",
			Short: "Run a container",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.Run(args)
			},
		},
		&cobra.Command{
			Use:   "pull NAME",
			Short: "Pull an image from docker hub",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.Pull(args)
			},
		},
		&cobra.Command{
			Use:   "rm NAME",
			Short: "Remove an image locally",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.Remove(args)
			},
		},
		&cobra.Command{
			Use:   "ls",
			Short: "List local images",
			RunE: func(cmd *cobra.Command, args []string) error {
				return command.Ls(args)
			},
		},
	}

	for _, cmd := range cmdList {
		rootCmd.AddCommand(cmd)
	}
	return rootCmd
}

// Execute runs cobra command
func Execute(cmd *cobra.Command) error {
	if err := config.Init(); err != nil {
		return err
	}
	if err := cmd.Execute(); err != nil {
		return err
	}
	return nil
}
