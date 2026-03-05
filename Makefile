# Slap Makefile
# Build and test the slap CLI tool and demo app.
#
# This Makefile leverages Go's built-in build cache for .go file tracking.
# Make handles orchestration; Go handles rebuild decisions.
#
# Usage:
#   make          Build all binaries
#   make test     Run tests
#   make help     Show all targets

# =============================================================================
# Shell Configuration
# =============================================================================

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules

# =============================================================================
# Variables
# =============================================================================

# Build output directory.
BIN := bin

# All packages under the module.
PKGS := ./...

# Version from VERSION file.
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")

# Build metadata (ISO 8601 datetime, can be overridden by CI).
BUILD_METADATA ?= local:$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Full version with build metadata (semver format).
FULL_VERSION := $(VERSION)+$(BUILD_METADATA)

# Release build mode (strips symbols and debug info for smaller binaries).
RELEASE ?= false

# Format check mode (FMT_CHECK=true verifies formatting without writing).
FMT_CHECK ?= false

# Go build flags with version injection.
# Release mode adds -s (strip symbol table) and -w (strip DWARF debug info).
ifeq ($(RELEASE),true)
LDFLAGS := -ldflags "-s -w -X main.Version=$(FULL_VERSION)"
else
LDFLAGS := -ldflags "-X main.Version=$(FULL_VERSION)"
endif

# Test flags: -count=1 disables caching for explicit test runs.
TEST_FLAGS := -count=1

# =============================================================================
# Default target
# =============================================================================

.DEFAULT_GOAL := all

.PHONY: all
all: build ## Build all binaries (default)

# =============================================================================
# Help
# =============================================================================

.PHONY: help
help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# =============================================================================
# Build targets
# =============================================================================

# Build targets are PHONY — Go's build cache handles rebuild decisions.
# Make's job is orchestration, not .go file dependency tracking.

.PHONY: build
build: build-slap build-demo ## Compile all binaries to bin/

.PHONY: build-slap
build-slap: | $(BIN)
	go build $(LDFLAGS) -o $(BIN)/slap ./cmd/slap/

.PHONY: build-demo
build-demo: | $(BIN)
	go build $(LDFLAGS) -o $(BIN)/demo ./cmd/demo/

# -----------------------------------------------------------------------------
# Cross-compilation targets (per-platform, both binaries)
# -----------------------------------------------------------------------------

.PHONY: build-darwin-amd64
build-darwin-amd64: | $(BIN) ## Cross-compile for macOS Intel
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/slap-darwin-amd64 ./cmd/slap/
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/demo-darwin-amd64 ./cmd/demo/

.PHONY: build-darwin-arm64
build-darwin-arm64: | $(BIN) ## Cross-compile for macOS Apple Silicon
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN)/slap-darwin-arm64 ./cmd/slap/
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN)/demo-darwin-arm64 ./cmd/demo/

.PHONY: build-linux-amd64
build-linux-amd64: | $(BIN) ## Cross-compile for Linux x64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/slap-linux-amd64 ./cmd/slap/
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN)/demo-linux-amd64 ./cmd/demo/

.PHONY: build-linux-arm64
build-linux-arm64: | $(BIN) ## Cross-compile for Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN)/slap-linux-arm64 ./cmd/slap/
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN)/demo-linux-arm64 ./cmd/demo/

.PHONY: build-all
build-all: build-darwin-amd64 build-darwin-arm64 build-linux-amd64 build-linux-arm64 ## Build all platform binaries (use -j for parallel)

# =============================================================================
# Test targets
# =============================================================================

.PHONY: test
test: vet ## Run tests
	go test $(TEST_FLAGS) $(PKGS)

# =============================================================================
# Quality targets
# =============================================================================

.PHONY: fmt
fmt: ## Format Go source files (FMT_CHECK=true to verify only)
ifeq ($(FMT_CHECK),true)
	test -z "$$(gofmt -l .)" || (echo "Files not formatted:" && gofmt -l . && exit 1)
else
	go fmt $(PKGS)
endif

.PHONY: vet
vet: ## Run go vet on all packages
	go vet $(PKGS)

.PHONY: lint
lint: fmt vet ## Run fmt and vet

# =============================================================================
# Maintenance targets
# =============================================================================

.PHONY: clean
clean: ## Remove build artifacts and clear test cache
	rm -rf $(BIN)
	go clean -testcache

.PHONY: version
version: ## Display version and build settings
	@echo "Version:  $(VERSION)"
	@echo "Full:     $(FULL_VERSION)"
	@echo "Release:  $(RELEASE)"

# =============================================================================
# Directory creation (order-only prerequisites)
# =============================================================================

$(BIN):
	mkdir -p $(BIN)
