package commands

import (
	"context"
	"fmt"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssh <handle>",
		Short: "Open a shell into a sandbox",
		Long:  "Exec into the sandbox pod for the given handle.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			handle := args[0]

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			client, err := kube.NewClient()
			if err != nil {
				return fmt.Errorf("connecting to cluster: %w", err)
			}

			ns := cfg.Namespace
			name := "sandbox-" + handle

			claimGVR := schema.GroupVersionResource{
				Group:    "agentsandbox.dev",
				Version:  "v1",
				Resource: "sandboxclaims",
			}

			claim, err := client.Dynamic().Resource(claimGVR).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("getting SandboxClaim %q: %w", name, err)
			}

			podName := extractPodName(claim.Object)
			if podName == "-" || podName == "" {
				return fmt.Errorf("sandbox %q does not have a pod assigned yet", handle)
			}

			fmt.Printf("connecting to pod %s...\n", podName)
			return kube.Exec(ns, podName, []string{"/bin/sh"})
		},
	}

	return cmd
}
