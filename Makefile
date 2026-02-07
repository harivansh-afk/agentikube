.PHONY: build install clean fmt vet lint

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
