.PHONY: build test fmt vet check

build:
	go build ./...

test:
	go test ./...

fmt:
	gofmt -l .

vet:
	go vet ./...

# Baseline check before commit/merge (Development Workflow in the constitution).
# golangci-lint is not yet installed in this repo; add it here if/when it is.
check: fmt vet test
