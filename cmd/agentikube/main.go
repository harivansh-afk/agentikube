package main

import (
	"fmt"
	"os"

	"github.com/rathi/agentikube/internal/commands"
	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	rootCmd := &cobra.Command{
		Use:   "agentikube",
		Short: "CLI for long-running agent sandboxes on Kubernetes",
		Long:  "agentikube provisions and manages long-running agent sandboxes on AWS using Kubernetes.",
	}

	rootCmd.PersistentFlags().String("config", "agentikube.yaml", "path to config file")

	rootCmd.AddCommand(
		commands.NewInitCmd(),
		commands.NewUpCmd(),
		commands.NewCreateCmd(),
		commands.NewListCmd(),
		commands.NewSSHCmd(),
		commands.NewDownCmd(),
		commands.NewDestroyCmd(),
		commands.NewStatusCmd(),
	)

	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
