package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/api/errors"
)

func NewDestroyCmd() *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "destroy <handle>",
		Short: "Destroy a sandbox and its resources",
		Long:  "Deletes the SandboxClaim, Secret, and PVC for the given handle.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			handle := args[0]

			if !yes {
				fmt.Printf("are you sure you want to destroy sandbox %q? [y/N] ", handle)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
				if answer != "y" && answer != "yes" {
					fmt.Println("aborted")
					return nil
				}
			}

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

			secretGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
			pvcGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaims"}

			// Delete SandboxClaim
			err = client.Dynamic().Resource(claimGVR).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("deleting SandboxClaim %q: %w", name, err)
			}
			fmt.Printf("[ok] SandboxClaim %q deleted\n", name)

			// Delete Secret
			err = client.Dynamic().Resource(secretGVR).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil {
				return fmt.Errorf("deleting Secret %q: %w", name, err)
			}
			fmt.Printf("[ok] Secret %q deleted\n", name)

			// Delete PVC (best-effort)
			err = client.Dynamic().Resource(pvcGVR).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
			if err != nil {
				if !errors.IsNotFound(err) {
					fmt.Printf("[warn] could not delete PVC %q: %v\n", name, err)
				}
			} else {
				fmt.Printf("[ok] PVC %q deleted\n", name)
			}

			fmt.Printf("\nsandbox %q destroyed\n", handle)
			return nil
		},
	}

	cmd.Flags().BoolVar(&yes, "yes", false, "skip confirmation prompt")

	return cmd
}
