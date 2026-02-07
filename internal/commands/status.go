package commands

import (
	"context"
	"fmt"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cluster and sandbox status",
		Long:  "Displays warm pool status, sandbox counts, and compute node information.",
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

			// Warm pool status
			wp, err := client.Dynamic().Resource(sandboxWarmPoolGVR).Namespace(ns).Get(ctx, "sandbox-warm-pool", metav1.GetOptions{})
			if err != nil {
				fmt.Printf("warm pool: not found (%v)\n", err)
			} else {
				spec, _ := wp.Object["spec"].(map[string]interface{})
				status, _ := wp.Object["status"].(map[string]interface{})

				replicas := getInt64(spec, "replicas")
				readyReplicas := getInt64(status, "readyReplicas")
				pendingReplicas := getInt64(status, "pendingReplicas")

				fmt.Println("warm pool:")
				fmt.Printf("  desired:  %d\n", replicas)
				fmt.Printf("  ready:    %d\n", readyReplicas)
				fmt.Printf("  pending:  %d\n", pendingReplicas)
			}

			// Sandbox count
			claims, err := client.Dynamic().Resource(sandboxClaimGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
			if err != nil {
				fmt.Printf("\nsandboxes: error listing (%v)\n", err)
			} else {
				fmt.Printf("\nsandboxes: %d\n", len(claims.Items))
			}

			// Karpenter nodes (if applicable)
			if cfg.Compute.Type == "karpenter" {
				nodes, err := client.Clientset().CoreV1().Nodes().List(ctx, metav1.ListOptions{
					LabelSelector: "karpenter.sh/nodepool",
				})
				if err != nil {
					fmt.Printf("\nkarpenter nodes: error listing (%v)\n", err)
				} else {
					fmt.Printf("\nkarpenter nodes: %d\n", len(nodes.Items))
				}
			}

			return nil
		},
	}

	return cmd
}

func getInt64(m map[string]interface{}, key string) int64 {
	if m == nil {
		return 0
	}
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int64:
		return n
	case float64:
		return int64(n)
	case int:
		return int64(n)
	default:
		return 0
	}
}
