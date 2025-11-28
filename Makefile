.PHONY: help test test-unit test-corpus test-all lint build clean fmt vet tidy sync examples

LINTER = "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2"

# Canonical doterr.go location
CANONICAL := $(HOME)/Projects/go-pkgs/go-doterr/doterr.go

# Directory to sync (default to current directory)
DIR ?= .

# Default target
help:
	@echo "Available targets:"
	@echo "  make test         - Run unit tests"
	@echo "  make test-corpus  - Run fuzz corpus regression tests"
	@echo "  make test-all     - Run all tests (unit + corpus)"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make fmt          - Format code with gofmt"
	@echo "  make vet          - Run go vet"
	@echo "  make tidy         - Run go mod tidy"
	@echo "  make build        - Build the package"
	@echo "  make examples     - Build all examples to ./bin/"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make ci           - Run all CI checks (fmt, vet, lint, test-all)"
	@echo "  make sync DIR=<path> [DRY_RUN=1] - Sync doterr.go to all subdirectories"

# Go environment
GO := go

# Run unit tests
test: test-unit

test-unit:
	cd test && $(GO) test -v -race -coverprofile=./coverage.txt -covermode=atomic ./...

# Run fuzz corpus regression tests
test-corpus:
	cd test && $(GO) test -v -run=TestFuzzCorpus

# Run all tests
test-all: test-unit test-corpus

# Run linter
lint:
	go run $(LINTER) run ./... --timeout=5m

# Format code
fmt:
	gofmt -s -w .

# Run go vet
vet:
	$(GO) vet ./...

# Run go mod tidy
tidy:
	$(GO) mod tidy
	cd test && $(GO) mod tidy

# Build the package
build:
	$(GO) build ./...

# Build all examples to ./bin/
examples:
	@mkdir -p bin
	@for example in examples/*/; do \
		name=$$(basename $$example); \
		echo "Building $$name..."; \
		cd $$example && \
		go mod init example 2>/dev/null || true && \
		go mod edit -replace github.com/mikeschinkel/go-doterr=../.. && \
		go mod tidy && \
		go build -o ../../bin/$$name . && \
		cd ../..; \
	done
	@echo "Examples built to ./bin/"

# Clean build artifacts
clean:
	$(GO) clean
	rm -f coverage.txt
	rm -rf bin
	cd test && $(GO) clean

# Run all CI checks locally
ci: fmt vet lint test-all
	@echo "All CI checks passed!"

# Sync doterr.go to all subdirectories
# Usage: make sync DIR=<path> [DRY_RUN=1]
sync:
	@if [ "$(DIR)" = "." ]; then \
		echo "Error: DIR not specified. Usage: make sync DIR=<path>"; \
		echo "Example: make sync DIR=$$HOME/Projects/xmlui"; \
		exit 1; \
	fi
	@EXPANDED_DIR=$$(eval echo "$(DIR)") && cd "$$EXPANDED_DIR" && bash $(CURDIR)/scripts/sync.sh
