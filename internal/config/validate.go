package config

import (
	"fmt"
	"strings"
)

// Validate checks that all required fields are present and values are valid.
func Validate(cfg *Config) error {
	var errs []string

	if cfg.Namespace == "" {
		errs = append(errs, "namespace is required")
	}

	// Compute validation
	switch cfg.Compute.Type {
	case "karpenter":
		if len(cfg.Compute.InstanceTypes) == 0 {
			errs = append(errs, "compute.instanceTypes is required when type is karpenter")
		}
		if len(cfg.Compute.CapacityTypes) == 0 {
			errs = append(errs, "compute.capacityTypes is required when type is karpenter")
		}
		if cfg.Compute.MaxCPU <= 0 {
			errs = append(errs, "compute.maxCpu must be > 0")
		}
		if cfg.Compute.MaxMemory == "" {
			errs = append(errs, "compute.maxMemory is required when type is karpenter")
		}
	case "fargate":
		if len(cfg.Compute.FargateSelectors) == 0 {
			errs = append(errs, "compute.fargateSelectors is required when type is fargate")
		}
	case "":
		errs = append(errs, "compute.type is required (karpenter or fargate)")
	default:
		errs = append(errs, fmt.Sprintf("compute.type must be karpenter or fargate, got %q", cfg.Compute.Type))
	}

	// Storage validation
	if cfg.Storage.Type == "" {
		errs = append(errs, "storage.type is required")
	} else if cfg.Storage.Type != "efs" {
		errs = append(errs, fmt.Sprintf("storage.type must be efs, got %q", cfg.Storage.Type))
	}
	if cfg.Storage.FilesystemID == "" {
		errs = append(errs, "storage.filesystemId is required")
	}
	if cfg.Storage.BasePath == "" {
		errs = append(errs, "storage.basePath is required")
	}
	if cfg.Storage.ReclaimPolicy == "" {
		cfg.Storage.ReclaimPolicy = "Retain"
	} else if cfg.Storage.ReclaimPolicy != "Retain" && cfg.Storage.ReclaimPolicy != "Delete" {
		errs = append(errs, fmt.Sprintf("storage.reclaimPolicy must be Retain or Delete, got %q", cfg.Storage.ReclaimPolicy))
	}

	// Storage defaults
	if cfg.Storage.UID == 0 {
		cfg.Storage.UID = 1000
	}
	if cfg.Storage.GID == 0 {
		cfg.Storage.GID = 1000
	}

	// Sandbox validation
	if cfg.Sandbox.Image == "" {
		errs = append(errs, "sandbox.image is required")
	}
	if len(cfg.Sandbox.Ports) == 0 {
		errs = append(errs, "sandbox.ports is required")
	}
	if cfg.Sandbox.MountPath == "" {
		errs = append(errs, "sandbox.mountPath is required")
	}

	// Warm pool defaults
	if cfg.Sandbox.WarmPool.Size == 0 && cfg.Sandbox.WarmPool.Enabled {
		cfg.Sandbox.WarmPool.Size = 5
	}
	if cfg.Sandbox.WarmPool.TTLMinutes == 0 {
		cfg.Sandbox.WarmPool.TTLMinutes = 120
	}

	// Probes defaults
	if cfg.Sandbox.Probes.Port == 0 && len(cfg.Sandbox.Ports) > 0 {
		cfg.Sandbox.Probes.Port = cfg.Sandbox.Ports[0]
	}
	if cfg.Sandbox.Probes.StartupFailureThreshold == 0 {
		cfg.Sandbox.Probes.StartupFailureThreshold = 30
	}

	if len(errs) > 0 {
		return fmt.Errorf("config validation errors:\n  - %s", strings.Join(errs, "\n  - "))
	}

	return nil
}
