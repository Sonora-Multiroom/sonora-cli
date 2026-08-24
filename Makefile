.PHONY: build test fmt vet check

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X sonora-cli/internal/version.Version=$(VERSION)" -o sonora.exe ./cmd/sonora

test:
	go test ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

# Baseline check before commit/merge (Development Workflow in the constitution).
# golangci-lint is not yet installed in this repo; add it here if/when it is.
check: fmt vet test
