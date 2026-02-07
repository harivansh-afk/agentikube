package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all sandboxes",
		Long:  "Lists all SandboxClaims in the configured namespace.",
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

			claimGVR := schema.GroupVersionResource{
				Group:    "agentsandbox.dev",
				Version:  "v1",
				Resource: "sandboxclaims",
			}

			list, err := client.Dynamic().Resource(claimGVR).Namespace(cfg.Namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("listing SandboxClaims: %w", err)
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "HANDLE\tSTATUS\tAGE\tPOD")

			for _, item := range list.Items {
				name := item.GetName()
				handle := name
				if len(name) > 8 && name[:8] == "sandbox-" {
					handle = name[8:]
				}

				status := extractStatus(item.Object)
				podName := extractPodName(item.Object)
				age := formatAge(item.GetCreationTimestamp().Time)

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", handle, status, age, podName)
			}

			w.Flush()
			return nil
		},
	}

	return cmd
}

func extractStatus(obj map[string]interface{}) string {
	status, ok := obj["status"].(map[string]interface{})
	if !ok {
		return "Unknown"
	}

	conditions, ok := status["conditions"].([]interface{})
	if !ok || len(conditions) == 0 {
		return "Pending"
	}

	// Look for the Ready condition
	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		condType, _ := cond["type"].(string)
		condStatus, _ := cond["status"].(string)
		if condType == "Ready" {
			if condStatus == "True" {
				return "Ready"
			}
			reason, _ := cond["reason"].(string)
			if reason != "" {
				return reason
			}
			return "NotReady"
		}
	}

	return "Pending"
}

func extractPodName(obj map[string]interface{}) string {
	status, ok := obj["status"].(map[string]interface{})
	if ok {
		if podName, ok := status["podName"].(string); ok && podName != "" {
			return podName
		}
	}

	// Fall back to annotations
	metadata, ok := obj["metadata"].(map[string]interface{})
	if ok {
		annotations, ok := metadata["annotations"].(map[string]interface{})
		if ok {
			if podName, ok := annotations["agentsandbox.dev/pod-name"].(string); ok {
				return podName
			}
		}
	}

	return "-"
}

func formatAge(created time.Time) string {
	d := time.Since(created)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%ds", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
}
