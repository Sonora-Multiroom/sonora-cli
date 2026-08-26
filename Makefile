.PHONY: build docker-build test fmt vet check

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags "-X sonora-cli/internal/version.Version=$(VERSION)" -o sonora.exe ./cmd/sonora

# Builds the Linux binary inside a Go container, for devs without a local Go toolchain.
docker-build:
	docker run --rm -v $(CURDIR):/app -w /app golang:1.27-alpine \
		go build -ldflags "-X sonora-cli/internal/version.Version=$(VERSION)" -o sonora ./cmd/sonora

test:
	go test ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

# Baseline check before commit/merge (Development Workflow in the constitution).
# golangci-lint is not yet installed in this repo; add it here if/when it is.
check: fmt vet test
