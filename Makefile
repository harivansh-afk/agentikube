.PHONY: build install clean fmt vet lint crds helm-lint helm-template

build:
	go build -o agentikube ./cmd/agentikube

install:
	go install ./cmd/agentikube

clean:
	rm -f agentikube

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

crds:
	./scripts/download-crds.sh

helm-lint:
	helm lint chart/agentikube/

helm-template:
	helm template agentikube chart/agentikube/ \
		--namespace sandboxes \
		--set storage.filesystemId=fs-test \
		--set sandbox.image=test:latest \
		--set compute.clusterName=test-cluster
