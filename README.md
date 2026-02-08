# agentikube

[![Go Version](https://img.shields.io/github/go-mod/go-version/harivansh-afk/agentikube)](https://github.com/harivansh-afk/agentikube/blob/main/go.mod)
[![Helm Version](https://img.shields.io/badge/helm%20chart-0.1.0-blue)](https://github.com/harivansh-afk/agentikube/tree/main/chart/agentikube)
[![Release](https://img.shields.io/github/v/release/harivansh-afk/agentikube)](https://github.com/harivansh-afk/agentikube/releases/latest)

Isolated stateful agent sandboxes on Kubernetes

<img width="1023" height="745" alt="image" src="https://github.com/user-attachments/assets/d62b6d99-b6bf-4ac3-9fb3-9b8373afbbec" />

## Install

```bash
helm install agentikube oci://ghcr.io/harivansh-afk/agentikube \
  -n sandboxes --create-namespace \
  -f my-values.yaml
```

Create a `my-values.yaml` with your cluster details:

```yaml
compute:
  clusterName: my-eks-cluster
storage:
  filesystemId: fs-0123456789abcdef0
sandbox:
  image: my-registry/sandbox:latest
```

See [`values.yaml`](chart/agentikube/values.yaml) for all options.

## CLI

The Go CLI handles runtime operations that are inherently imperative:

```bash
agentikube create demo --provider openai --api-key <key>
agentikube list
agentikube ssh demo
agentikube status
agentikube destroy demo
```

Build it with `go build ./cmd/agentikube` or `make build`.

## What gets created

The Helm chart installs:

- StorageClass (`efs-sandbox`) backed by your EFS filesystem
- SandboxTemplate defining the pod spec
- NetworkPolicy for ingress/egress rules
- SandboxWarmPool (optional, enabled by default)
- Karpenter NodePool + EC2NodeClass (optional, when `compute.type: karpenter`)

Each `agentikube create <handle>` then adds a Secret, SandboxClaim, and workspace PVC for that user.

## Project layout

```
cmd/agentikube/              CLI entrypoint
internal/                    config, manifest rendering, kube helpers
chart/agentikube/            Helm chart
scripts/                     CRD download helper
```

## Development

```bash
make build                   # compile CLI
make helm-lint               # lint the chart
make helm-template           # dry-run render
go test ./...                # run tests
```

## Good to know

- Storage is EFS-only for now
- `kubectl` must be installed (used by `init` and `ssh`)
- Fargate is validated in config but templates only cover Karpenter so far
- [k9s](https://k9scli.io/) is great for browsing sandbox resources

## Context

[![Blog Post: Isolated Long-Running Agents with Kubernetes](https://hari.tech/thoughts/isolated-long-running-agents-with-kubernetes/opengraph-image?5c0605812d5fdbb7)](https://hari.tech/thoughts/isolated-long-running-agents-with-kubernetes)
