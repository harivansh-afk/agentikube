package manifest

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/rathi/agentikube/internal/config"
)

// Generate renders all applicable Kubernetes manifests from the embedded
// templates using the provided configuration. Templates are selected based
// on the compute type and warm pool settings.
func Generate(cfg *config.Config) ([]byte, error) {
	tmpl, err := template.ParseFS(templateFS, "templates/*.yaml.tmpl")
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	// Always-rendered templates
	names := []string{
		"namespace.yaml.tmpl",
		"storageclass-efs.yaml.tmpl",
		"sandbox-template.yaml.tmpl",
	}

	// Conditionally add Karpenter templates
	if cfg.Compute.Type == "karpenter" {
		names = append(names,
			"karpenter-nodepool.yaml.tmpl",
			"karpenter-ec2nodeclass.yaml.tmpl",
		)
	}

	// Conditionally add warm pool template
	if cfg.Sandbox.WarmPool.Enabled {
		names = append(names, "warm-pool.yaml.tmpl")
	}

	var out bytes.Buffer
	for i, name := range names {
		if i > 0 {
			out.WriteString("---\n")
		}
		if err := tmpl.ExecuteTemplate(&out, name, cfg); err != nil {
			return nil, fmt.Errorf("rendering template %s: %w", name, err)
		}
	}

	return out.Bytes(), nil
}
