package commands

import (
	"context"
	"fmt"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/rathi/agentikube/internal/manifest"
	"github.com/spf13/cobra"
)

func NewUpCmd() *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Apply sandbox infrastructure to the cluster",
		Long:  "Generates and applies all sandbox manifests (templates, warm pool, storage, compute).",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			manifests, err := manifest.Generate(cfg)
			if err != nil {
				return fmt.Errorf("generating manifests: %w", err)
			}

			if dryRun {
				fmt.Print(string(manifests))
				return nil
			}

			client, err := kube.NewClient()
			if err != nil {
				return fmt.Errorf("connecting to cluster: %w", err)
			}

			if err := client.ServerSideApply(ctx, manifests); err != nil {
				return fmt.Errorf("applying manifests: %w", err)
			}
			fmt.Println("[ok] manifests applied")

			if cfg.Sandbox.WarmPool.Enabled {
				fmt.Println("waiting for warm pool to become ready...")
				if err := client.WaitForReady(ctx, cfg.Namespace, "sandboxwarmpools", "sandbox-warm-pool"); err != nil {
					return fmt.Errorf("waiting for warm pool: %w", err)
				}
				fmt.Println("[ok] warm pool ready")
			}

			fmt.Println("\ninfrastructure is up")
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "print manifests to stdout without applying")

	return cmd
}
