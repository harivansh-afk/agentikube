package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration parsed from agentikube.yaml.
type Config struct {
	Namespace string         `yaml:"namespace"`
	Compute   ComputeConfig  `yaml:"compute"`
	Storage   StorageConfig  `yaml:"storage"`
	Sandbox   SandboxConfig  `yaml:"sandbox"`
}

type ComputeConfig struct {
	Type             string            `yaml:"type"` // karpenter | fargate
	InstanceTypes    []string          `yaml:"instanceTypes"`
	CapacityTypes    []string          `yaml:"capacityTypes"`
	MaxCPU           int               `yaml:"maxCpu"`
	MaxMemory        string            `yaml:"maxMemory"`
	Consolidation    bool              `yaml:"consolidation"`
	FargateSelectors []FargateSelector `yaml:"fargateSelectors"`
}

type FargateSelector struct {
	Namespace string `yaml:"namespace"`
}

type StorageConfig struct {
	Type          string `yaml:"type"` // efs
	FilesystemID  string `yaml:"filesystemId"`
	BasePath      string `yaml:"basePath"`
	UID           int    `yaml:"uid"`
	GID           int    `yaml:"gid"`
	ReclaimPolicy string `yaml:"reclaimPolicy"`
}

type SandboxConfig struct {
	Image           string            `yaml:"image"`
	Ports           []int             `yaml:"ports"`
	MountPath       string            `yaml:"mountPath"`
	Resources       ResourcesConfig   `yaml:"resources"`
	Env             map[string]string `yaml:"env"`
	SecurityContext SecurityContext   `yaml:"securityContext"`
	Probes          ProbesConfig      `yaml:"probes"`
	WarmPool        WarmPoolConfig    `yaml:"warmPool"`
	NetworkPolicy   NetworkPolicy     `yaml:"networkPolicy"`
}

type ResourcesConfig struct {
	Requests ResourceValues `yaml:"requests"`
	Limits   ResourceValues `yaml:"limits"`
}

type ResourceValues struct {
	CPU    string `yaml:"cpu"`
	Memory string `yaml:"memory"`
}

type SecurityContext struct {
	RunAsUser    int  `yaml:"runAsUser"`
	RunAsGroup   int  `yaml:"runAsGroup"`
	RunAsNonRoot bool `yaml:"runAsNonRoot"`
}

type ProbesConfig struct {
	Port                    int `yaml:"port"`
	StartupFailureThreshold int `yaml:"startupFailureThreshold"`
}

type WarmPoolConfig struct {
	Enabled    bool `yaml:"enabled"`
	Size       int  `yaml:"size"`
	TTLMinutes int  `yaml:"ttlMinutes"`
}

type NetworkPolicy struct {
	EgressAllowAll bool  `yaml:"egressAllowAll"`
	IngressPorts   []int `yaml:"ingressPorts"`
}

// Load reads and parses the config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := Validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}
