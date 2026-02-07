# agentikube

[![Go Version](https://img.shields.io/github/go-mod/go-version/harivansh-afk/agentikube)](https://github.com/harivansh-afk/agentikube/blob/main/go.mod)
[![Release](https://img.shields.io/github/v/release/harivansh-afk/agentikube)](https://github.com/harivansh-afk/agentikube/releases/latest)

A small Go CLI that spins up isolated agent sandboxes on Kubernetes. Built for AWS setups (EFS + optional Karpenter).

<img width="1023" height="745" alt="image" src="https://github.com/user-attachments/assets/d62b6d99-b6bf-4ac3-9fb3-9b8373afbbec" />

## What it does

- **`init`** - Installs CRDs, checks prerequisites, ensures your namespace exists
- **`up`** - Renders and applies Kubernetes manifests from templates (`--dry-run` to preview)
- **`create <handle>`** - Spins up a sandbox for a user with provider credentials
- **`list`** - Shows all sandboxes with status, age, and pod name
- **`status`** - Warm pool numbers, sandbox count, Karpenter node count
- **`ssh <handle>`** - Drops you into a sandbox pod shell
- **`destroy <handle>`** - Tears down a single sandbox
- **`down`** - Removes shared infra but keeps existing user sandboxes

## Quick start

```bash
# 1. Copy and fill in your config
cp agentikube.example.yaml agentikube.yaml
# Edit: namespace, EFS filesystem ID, sandbox image, compute settings

# 2. Set things up
agentikube init
agentikube up

# 3. Create a sandbox and jump in
agentikube create demo --provider openai --api-key <key>
agentikube list
agentikube ssh demo
```

## What gets created

Running `up` applies these to your cluster:

- Namespace, StorageClass (`efs-sandbox`), SandboxTemplate
- Optionally: SandboxWarmPool, NodePool + EC2NodeClass (Karpenter)

Running `create <handle>` adds:

- A Secret and SandboxClaim per user
- A workspace PVC backed by EFS

## Project layout

```
cmd/agentikube/main.go         # entrypoint
internal/config/               # config structs + validation
internal/manifest/             # template rendering
internal/manifest/templates/   # k8s YAML templates
internal/kube/                 # kube client helpers
internal/commands/             # command implementations
agentikube.example.yaml        # example config
Makefile                       # build/install/fmt/vet
```

## Build and test locally

```bash
go build ./...
go test ./...
go run ./cmd/agentikube --help

# Smoke test manifest generation
./agentikube up --dry-run --config agentikube.example.yaml
```

## Good to know

- `storage.type` is `efs` only for now
- `kubectl` needs to be installed (used by `init` and `ssh`)
- Fargate is validated in config but templates only cover the Karpenter path so far
- No Go tests written yet - `go test` passes but reports no test files
- [k9s](https://k9scli.io/) is great for browsing sandbox resources (`brew install derailed/k9s/k9s`)
