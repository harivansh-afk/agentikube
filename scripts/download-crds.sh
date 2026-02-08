#!/usr/bin/env bash
set -euo pipefail

# Download agent-sandbox CRDs into chart/agentikube/crds/
# Run this before packaging the chart: make crds

REPO="kubernetes-sigs/agent-sandbox"
BRANCH="main"
BASE_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}/k8s/crds"
DEST="$(cd "$(dirname "$0")/.." && pwd)/chart/agentikube/crds"

CRDS=(
  sandboxtemplates.yaml
  sandboxclaims.yaml
  sandboxwarmpools.yaml
)

echo "Downloading CRDs from ${REPO}@${BRANCH} ..."
mkdir -p "$DEST"

for crd in "${CRDS[@]}"; do
  echo "  ${crd}"
  curl -sSfL "${BASE_URL}/${crd}" -o "${DEST}/${crd}"
done

echo "CRDs written to ${DEST}"
