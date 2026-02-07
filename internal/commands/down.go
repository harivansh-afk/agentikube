package commands

import (
	"context"
	"fmt"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Remove sandbox infrastructure (preserves user sandboxes)",
		Long:  "Deletes the SandboxWarmPool and SandboxTemplate. User sandboxes are preserved.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			client, err := kube.NewClient()
			if err != nil {
				return fmt.Errorf("connecting to cluster: %w", err)
			}

			ns := cfg.Namespace

			warmPoolGVR := schema.GroupVersionResource{
				Group:    "agentsandbox.dev",
				Version:  "v1",
				Resource: "sandboxwarmpools",
			}

			templateGVR := schema.GroupVersionResource{
				Group:    "agentsandbox.dev",
				Version:  "v1",
				Resource: "sandboxtemplates",
			}

			err = client.Dynamic().Resource(warmPoolGVR).Namespace(ns).Delete(ctx, "sandbox-warm-pool", metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("[warn] could not delete SandboxWarmPool: %v\n", err)
			} else {
				fmt.Println("[ok] SandboxWarmPool deleted")
			}

			err = client.Dynamic().Resource(templateGVR).Namespace(ns).Delete(ctx, "sandbox-template", metav1.DeleteOptions{})
			if err != nil {
				fmt.Printf("[warn] could not delete SandboxTemplate: %v\n", err)
			} else {
				fmt.Println("[ok] SandboxTemplate deleted")
			}

			fmt.Println("\nwarm pool and template deleted. User sandboxes are preserved.")
			return nil
		},
	}

	return cmd
}
