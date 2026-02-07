package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewCreateCmd() *cobra.Command {
	var provider string
	var apiKey string

	cmd := &cobra.Command{
		Use:   "create <handle>",
		Short: "Create a new sandbox for an agent",
		Long:  "Creates a Secret and SandboxClaim for the given handle, then waits for it to be ready.",
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

			// Create the secret with provider credentials
			secret := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Secret",
					"metadata": map[string]interface{}{
						"name":      name,
						"namespace": ns,
					},
					"stringData": map[string]interface{}{
						"PROVIDER":     provider,
						"PROVIDER_KEY": apiKey,
						"USER_NAME":    handle,
					},
				},
			}

			secretGVR := coreGVR("secrets")
			_, err = client.Dynamic().Resource(secretGVR).Namespace(ns).Create(ctx, secret, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("creating secret %q: %w", name, err)
			}
			fmt.Printf("[ok] secret %q created\n", name)

			// Create the SandboxClaim
			claim := &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "extensions.agents.x-k8s.io/v1alpha1",
					"kind":       "SandboxClaim",
					"metadata": map[string]interface{}{
						"name":      name,
						"namespace": ns,
					},
					"spec": map[string]interface{}{
						"templateRef": map[string]interface{}{
							"name": "sandbox-template",
						},
						"secretRef": map[string]interface{}{
							"name": name,
						},
					},
				},
			}

			_, err = client.Dynamic().Resource(sandboxClaimGVR).Namespace(ns).Create(ctx, claim, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("creating SandboxClaim %q: %w", name, err)
			}
			fmt.Printf("[ok] SandboxClaim %q created\n", name)

			// Wait for the sandbox to become ready (3 min timeout)
			fmt.Println("waiting for sandbox to be ready...")
			waitCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
			defer cancel()

			if err := client.WaitForReady(waitCtx, ns, sandboxClaimGVR, name); err != nil {
				return fmt.Errorf("waiting for sandbox: %w", err)
			}

			fmt.Printf("\nsandbox %q is ready\n", handle)
			fmt.Printf("  name:      %s\n", name)
			fmt.Printf("  namespace: %s\n", ns)
			fmt.Printf("  ssh:       agentikube ssh %s\n", handle)
			return nil
		},
	}

	cmd.Flags().StringVar(&provider, "provider", "", "LLM provider name (env: SANDBOX_LLM_PROVIDER)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "LLM provider API key (env: SANDBOX_API_KEY)")

	return cmd
}
