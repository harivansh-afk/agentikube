package commands

import (
	"github.com/rathi/agentikube/internal/config"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	sandboxClaimGVR = schema.GroupVersionResource{
		Group:    "extensions.agents.x-k8s.io",
		Version:  "v1alpha1",
		Resource: "sandboxclaims",
	}
	sandboxTemplateGVR = schema.GroupVersionResource{
		Group:    "extensions.agents.x-k8s.io",
		Version:  "v1alpha1",
		Resource: "sandboxtemplates",
	}
	sandboxWarmPoolGVR = schema.GroupVersionResource{
		Group:    "extensions.agents.x-k8s.io",
		Version:  "v1alpha1",
		Resource: "sandboxwarmpools",
	}
)

func coreGVR(resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{Group: "", Version: "v1", Resource: resource}
}

func loadConfig(cmd *cobra.Command) (*config.Config, error) {
	cfgPath, _ := cmd.Flags().GetString("config")
	return config.Load(cfgPath)
}
