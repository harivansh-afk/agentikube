package chart_test

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// helmTemplate runs helm template with the given extra args and returns stdout.
func helmTemplate(t *testing.T, extraArgs ...string) string {
	t.Helper()
	args := []string{
		"template", "agentikube", "chart/agentikube/",
		"--namespace", "sandboxes",
		"--set", "storage.filesystemId=fs-test",
		"--set", "sandbox.image=test:latest",
		"--set", "compute.clusterName=test-cluster",
	}
	args = append(args, extraArgs...)
	cmd := exec.Command("helm", args...)
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm template failed: %v\n%s", err, out)
	}
	return string(out)
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// This test file lives at chart/agentikube_test.go, so repo root is ..
	return dir + "/.."
}

func TestHelmLint(t *testing.T) {
	cmd := exec.Command("helm", "lint", "chart/agentikube/")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("helm lint failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "0 chart(s) failed") {
		t.Fatalf("helm lint reported failures:\n%s", out)
	}
}

func TestHelmTemplateDefaultValues(t *testing.T) {
	output := helmTemplate(t)

	expected := []string{
		"kind: StorageClass",
		"kind: SandboxTemplate",
		"kind: SandboxWarmPool",
		"kind: NodePool",
		"kind: EC2NodeClass",
		"kind: NetworkPolicy",
	}
	for _, want := range expected {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in rendered output", want)
		}
	}
}

func TestHelmTemplateLabels(t *testing.T) {
	output := helmTemplate(t)

	labels := []string{
		"helm.sh/chart: agentikube-0.1.0",
		"app.kubernetes.io/name: agentikube",
		"app.kubernetes.io/instance: agentikube",
		"app.kubernetes.io/managed-by: Helm",
		`app.kubernetes.io/version: "0.1.0"`,
	}
	for _, label := range labels {
		if !strings.Contains(output, label) {
			t.Errorf("expected label %q in rendered output", label)
		}
	}
}

func TestHelmTemplateKarpenterDisabled(t *testing.T) {
	output := helmTemplate(t, "--set", "compute.type=fargate")

	if strings.Contains(output, "kind: NodePool") {
		t.Error("NodePool should not be rendered when compute.type=fargate")
	}
	if strings.Contains(output, "kind: EC2NodeClass") {
		t.Error("EC2NodeClass should not be rendered when compute.type=fargate")
	}
	if !strings.Contains(output, "kind: StorageClass") {
		t.Error("StorageClass should always be rendered")
	}
	if !strings.Contains(output, "kind: SandboxTemplate") {
		t.Error("SandboxTemplate should always be rendered")
	}
}

func TestHelmTemplateWarmPoolDisabled(t *testing.T) {
	output := helmTemplate(t, "--set", "sandbox.warmPool.enabled=false")

	if strings.Contains(output, "kind: SandboxWarmPool") {
		t.Error("SandboxWarmPool should not be rendered when warmPool.enabled=false")
	}
	if !strings.Contains(output, "kind: SandboxTemplate") {
		t.Error("SandboxTemplate should always be rendered")
	}
}

func TestHelmTemplateEgressDisabled(t *testing.T) {
	output := helmTemplate(t,
		"--set", "sandbox.networkPolicy.egressAllowAll=false",
		"-s", "templates/networkpolicy.yaml",
	)

	if strings.Contains(output, "0.0.0.0/0") {
		t.Error("egress CIDR should not appear when egressAllowAll=false")
	}
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if strings.Contains(line, "policyTypes:") {
			block := strings.Join(lines[i:min(i+4, len(lines))], "\n")
			if strings.Contains(block, "Egress") {
				t.Error("Egress should not be in policyTypes when egressAllowAll=false")
			}
		}
	}
}

func TestHelmTemplateRequiredValues(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{
			name:    "missing filesystemId",
			args:    []string{"--set", "sandbox.image=test:latest", "--set", "compute.clusterName=test"},
			wantErr: "storage.filesystemId is required",
		},
		{
			name:    "missing sandbox image",
			args:    []string{"--set", "storage.filesystemId=fs-test", "--set", "compute.clusterName=test"},
			wantErr: "sandbox.image is required",
		},
		{
			name:    "missing clusterName for karpenter",
			args:    []string{"--set", "storage.filesystemId=fs-test", "--set", "sandbox.image=test:latest"},
			wantErr: "compute.clusterName is required for Karpenter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{
				"template", "agentikube", "chart/agentikube/",
				"--namespace", "sandboxes",
			}, tt.args...)
			cmd := exec.Command("helm", args...)
			cmd.Dir = repoRoot(t)
			out, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatal("expected helm template to fail for missing required value")
			}
			if !strings.Contains(string(out), tt.wantErr) {
				t.Errorf("expected error containing %q, got:\n%s", tt.wantErr, out)
			}
		})
	}
}

func TestHelmTemplateEnvVars(t *testing.T) {
	output := helmTemplate(t,
		"--set", "sandbox.env.MY_VAR=my-value",
		"-s", "templates/sandbox-template.yaml",
	)

	if !strings.Contains(output, "MY_VAR") {
		t.Error("expected MY_VAR in rendered env")
	}
	if !strings.Contains(output, "my-value") {
		t.Error("expected my-value in rendered env")
	}
}

func TestHelmTemplateNoEnvWhenEmpty(t *testing.T) {
	output := helmTemplate(t, "-s", "templates/sandbox-template.yaml")

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "env:" {
			t.Error("env: block should not appear when sandbox.env is empty")
		}
	}
}

func TestHelmTemplateNamespace(t *testing.T) {
	output := helmTemplate(t, "--namespace", "custom-ns")

	if !strings.Contains(output, "namespace: custom-ns") {
		t.Error("expected namespace: custom-ns in rendered output")
	}
}

func TestHelmTemplateConsolidationDisabled(t *testing.T) {
	output := helmTemplate(t,
		"--set", "compute.consolidation=false",
		"-s", "templates/karpenter-nodepool.yaml",
	)

	if !strings.Contains(output, "consolidationPolicy: WhenEmpty") {
		t.Error("expected consolidationPolicy: WhenEmpty when consolidation=false")
	}
	if strings.Contains(output, "WhenEmptyOrUnderutilized") {
		t.Error("should not have WhenEmptyOrUnderutilized when consolidation=false")
	}
}
