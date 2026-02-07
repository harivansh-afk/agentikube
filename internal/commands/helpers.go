package commands

import (
	"github.com/rathi/agentikube/internal/config"
	"github.com/spf13/cobra"
)

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	return config.Load(cfgPath)
}
