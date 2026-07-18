# ─── ax — TUI API Client ─────────────────────────────────────────────────────
# ──────────────────────────────────────────────────────────────────────────────

APP_NAME   := ax
CMD_DIR    := ./cmd/ax
BUILD_DIR  := ./build
BINARY     := $(BUILD_DIR)/$(APP_NAME)

# Detect OS for binary naming
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Windows_NT)
	BINARY := $(BUILD_DIR)/$(APP_NAME).exe
endif

# Build flags — version, commit, and date injected at link time
VERSION ?= $(shell git describe --tags 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: all build clean test test-short lint run fmt tidy help completion goreleaser-check

all: build

build: ## Build the binary
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BINARY) $(CMD_DIR)
	@echo "✓ Built $(BINARY)"

build-static: ## Build a fully static binary
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -tags netgo -trimpath -o $(BINARY) $(CMD_DIR)
	@echo "✓ Built static $(BINARY)"

clean: ## Remove build artifacts
	rm -rf $(BUILD_DIR) dist/
	@echo "✓ Cleaned"

test: ## Run all tests with race detector
	go test ./... -race -count=1 -v

test-short: ## Run tests without race detector (faster)
	go test ./... -count=1

test-pkg: ## Run tests for a specific package (usage: make test-pkg PKG=./internal/client/)
	go test $(PKG) -count=1 -v

lint: ## Run golangci-lint (must be installed)
	@which golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "⚠  golangci-lint not installed, skipping"

run: build ## Build and run
	@$(BINARY)

fmt: ## Format all Go code
	go fmt ./...

tidy: ## Tidy module dependencies
	go mod tidy
	go mod verify

vet: ## Run go vet
	go vet ./...

completion: ## Generate shell completion scripts
	@echo "=== Bash ==="
	@go run $(CMD_DIR) completion bash
	@echo ""
	@echo "=== Zsh ==="
	@go run $(CMD_DIR) completion zsh
	@echo ""
	@echo "=== Fish ==="
	@go run $(CMD_DIR) completion fish

goreleaser-check: ## Validate GoReleaser config and run a snapshot build
	@which goreleaser >/dev/null 2>&1 || (echo "⚠  goreleaser not installed, install from https://goreleaser.com/install/" && exit 1)
	goreleaser check -f .goreleaser.yaml
	GORELEASER_CURRENT_TAG=v0.0.0-test goreleaser release --snapshot --clean

help: ## Show this help
	@echo ''
	@echo '$(APP_NAME) — TUI API Client'
	@echo ''
	@echo 'Usage:'
	@echo '  make build            Build the binary'
	@echo '  make build-static     Build fully static binary'
	@echo '  make clean            Remove build artifacts'
	@echo '  make test             Run all tests (with race detector)'
	@echo '  make test-short       Run tests without race detector'
	@echo '  make test-pkg PKG=…   Run tests for a specific package'
	@echo '  make lint             Run golangci-lint'
	@echo '  make run              Build and run'
	@echo '  make fmt              Format Go code'
	@echo '  make vet              Run go vet'
	@echo '  make tidy             Tidy dependencies'
	@echo '  make completion       Generate shell completion scripts'
	@echo '  make goreleaser-check Validate GoReleaser config'
	@echo '  make help             Show this help'
	@echo ''
