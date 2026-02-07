package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rathi/agentikube/internal/kube"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const crdInstallURL = "https://raw.githubusercontent.com/agent-sandbox/agent-sandbox/main/deploy/install.yaml"

func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the cluster for agent sandboxes",
		Long:  "Checks prerequisites, installs CRDs, and creates the target namespace.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := loadConfig(cmd)
			if err != nil {
				return err
			}

			// Check kubectl context
			client, err := kube.NewClient()
			if err != nil {
				return fmt.Errorf("connecting to cluster: %w", err)
			}
			fmt.Println("[ok] connected to Kubernetes cluster")

			// Apply agent-sandbox CRDs
			fmt.Println("applying agent-sandbox CRDs...")
			out, err := exec.CommandContext(ctx, "kubectl", "apply", "-f", crdInstallURL).CombinedOutput()
			if err != nil {
				return fmt.Errorf("applying CRDs: %s: %w", strings.TrimSpace(string(out)), err)
			}
			fmt.Println("[ok] agent-sandbox CRDs applied")

			// Check for EFS CSI driver
			dsList, err := client.Clientset().AppsV1().DaemonSets("kube-system").List(ctx, metav1.ListOptions{})
			if err != nil {
				return fmt.Errorf("listing daemonsets in kube-system: %w", err)
			}
			efsFound := false
			for _, ds := range dsList.Items {
				if strings.Contains(ds.Name, "efs-csi") {
					efsFound = true
					break
				}
			}
			if efsFound {
				fmt.Println("[ok] EFS CSI driver found")
			} else {
				fmt.Println("[warn] EFS CSI driver not found - install it before using EFS storage")
			}

			// Check for Karpenter
			karpenterFound := false
			for _, ns := range []string{"karpenter", "kube-system"} {
				depList, err := client.Clientset().AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
				if err != nil {
					continue
				}
				for _, dep := range depList.Items {
					if strings.Contains(dep.Name, "karpenter") {
						karpenterFound = true
						break
					}
				}
				if karpenterFound {
					break
				}
			}
			if karpenterFound {
				fmt.Println("[ok] Karpenter found")
			} else {
				fmt.Println("[warn] Karpenter not found - required if compute.type is karpenter")
			}

			// Create namespace if it does not exist
			if err := client.EnsureNamespace(ctx, cfg.Namespace); err != nil {
				return fmt.Errorf("creating namespace %q: %w", cfg.Namespace, err)
			}
			fmt.Printf("[ok] namespace %q ready\n", cfg.Namespace)

			fmt.Println("\ninit complete")
			return nil
		},
	}

	return cmd
}
