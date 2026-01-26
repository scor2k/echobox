# Makefile for Echobox - SRE Interview Terminal

# Variables
BINARY_NAME=echobox
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags "-s -w"
BUILD_DIR=./build
CMD_DIR=./cmd/server
PKG_LIST=$(shell go list ./... | grep -v /vendor/)

# Docker variables (can be overridden)
DOCKER_PORT?=8080

# Default target
.DEFAULT_GOAL := help

# Colors for output
COLOR_RESET=\033[0m
COLOR_BOLD=\033[1m
COLOR_GREEN=\033[32m
COLOR_YELLOW=\033[33m
COLOR_BLUE=\033[34m

.PHONY: help
help: ## Show this help message
	@echo "$(COLOR_BOLD)Echobox - SRE Interview Terminal$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Usage:$(COLOR_RESET)"
	@echo "  make $(COLOR_GREEN)<target>$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Available targets:$(COLOR_RESET)"
	@awk 'BEGIN {FS = ":.*##"; printf ""} /^[a-zA-Z_-]+:.*?##/ { printf "  $(COLOR_GREEN)%-15s$(COLOR_RESET) %s\n", $$1, $$2 } /^##@/ { printf "\n$(COLOR_BOLD)%s$(COLOR_RESET)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: build
build: clean ## Build the binary
	@echo "$(COLOR_BLUE)Building $(BINARY_NAME)...$(COLOR_RESET)"
	@$(GO) build $(GOFLAGS) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Build complete: $(BINARY_NAME)$(COLOR_RESET)"

.PHONY: build-all
build-all: clean ## Build for all platforms (Linux, macOS, Windows)
	@echo "$(COLOR_BLUE)Building for all platforms...$(COLOR_RESET)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	@GOOS=linux GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	@GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	@GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	@echo "$(COLOR_GREEN)✓ Cross-compilation complete$(COLOR_RESET)"
	@ls -lh $(BUILD_DIR)

.PHONY: run
run: build ## Build and run the server
	@echo "$(COLOR_BLUE)Starting server...$(COLOR_RESET)"
	@mkdir -p sessions
	@OUTPUT_DIR=./sessions ./$(BINARY_NAME)

.PHONY: run-dev
run-dev: build ## Run with development settings
	@echo "$(COLOR_BLUE)Starting server in development mode...$(COLOR_RESET)"
	@mkdir -p sessions
	@CANDIDATE_NAME="dev_user" \
	 SESSION_TIMEOUT=3600 \
	 LOG_LEVEL=debug \
	 OUTPUT_DIR=./sessions \
	 ./$(BINARY_NAME)

.PHONY: watch
watch: ## Watch for changes and rebuild (requires entr: brew install entr)
	@echo "$(COLOR_YELLOW)Watching for changes... (Ctrl+C to stop)$(COLOR_RESET)"
	@find . -name "*.go" -o -name "*.html" -o -name "*.css" -o -name "*.js" | entr -r make run

##@ Testing

.PHONY: test
test: ## Run tests
	@echo "$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@$(GO) test -v -race -coverprofile=coverage.out $(PKG_LIST)
	@echo "$(COLOR_GREEN)✓ Tests complete$(COLOR_RESET)"

.PHONY: test-coverage
test-coverage: test ## Run tests with coverage report
	@echo "$(COLOR_BLUE)Generating coverage report...$(COLOR_RESET)"
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(COLOR_GREEN)✓ Coverage report: coverage.html$(COLOR_RESET)"

.PHONY: test-short
test-short: ## Run short tests only
	@echo "$(COLOR_BLUE)Running short tests...$(COLOR_RESET)"
	@$(GO) test -short $(PKG_LIST)

.PHONY: bench
bench: ## Run benchmarks
	@echo "$(COLOR_BLUE)Running benchmarks...$(COLOR_RESET)"
	@$(GO) test -bench=. -benchmem $(PKG_LIST)

##@ Code Quality

.PHONY: fmt
fmt: ## Format code
	@echo "$(COLOR_BLUE)Formatting code...$(COLOR_RESET)"
	@$(GO) fmt $(PKG_LIST)
	@echo "$(COLOR_GREEN)✓ Code formatted$(COLOR_RESET)"

.PHONY: vet
vet: ## Run go vet
	@echo "$(COLOR_BLUE)Running go vet...$(COLOR_RESET)"
	@$(GO) vet $(PKG_LIST)
	@echo "$(COLOR_GREEN)✓ Vet complete$(COLOR_RESET)"

.PHONY: lint
lint: ## Run golangci-lint (requires golangci-lint)
	@echo "$(COLOR_BLUE)Running linter...$(COLOR_RESET)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
		echo "$(COLOR_GREEN)✓ Lint complete$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW)⚠ golangci-lint not installed. Run: brew install golangci-lint$(COLOR_RESET)"; \
	fi

.PHONY: check
check: fmt vet ## Run all checks (fmt, vet)
	@echo "$(COLOR_GREEN)✓ All checks passed$(COLOR_RESET)"

##@ Dependencies

.PHONY: deps
deps: ## Download dependencies
	@echo "$(COLOR_BLUE)Downloading dependencies...$(COLOR_RESET)"
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Dependencies downloaded$(COLOR_RESET)"

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(COLOR_BLUE)Updating dependencies...$(COLOR_RESET)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ Dependencies updated$(COLOR_RESET)"

.PHONY: tidy
tidy: ## Tidy go.mod
	@echo "$(COLOR_BLUE)Tidying go.mod...$(COLOR_RESET)"
	@$(GO) mod tidy
	@echo "$(COLOR_GREEN)✓ go.mod tidied$(COLOR_RESET)"

##@ Docker

.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(COLOR_BLUE)Building Docker image...$(COLOR_RESET)"
	@docker build -t echobox:latest -t echobox:dev .
	@docker images echobox:latest
	@echo "$(COLOR_GREEN)✓ Docker image built: echobox:latest, echobox:dev$(COLOR_RESET)"

.PHONY: docker-build-prod
docker-build-prod: ## Build production Docker image with optimizations
	@echo "$(COLOR_BLUE)Building production Docker image...$(COLOR_RESET)"
	@docker build \
		--build-arg VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev") \
		-t echobox:prod \
		-t echobox:$(shell git rev-parse --short HEAD 2>/dev/null || echo "latest") \
		.
	@docker images echobox
	@echo "$(COLOR_GREEN)✓ Production image built$(COLOR_RESET)"

.PHONY: docker-run
docker-run: docker-build ## Build and run Docker container
	@echo "$(COLOR_BLUE)Starting Docker container...$(COLOR_RESET)"
	@mkdir -p sessions tasks
	@docker run -it --rm \
		-p $(DOCKER_PORT):8080 \
		-v $(PWD)/sessions:/output \
		-v $(PWD)/tasks:/tasks:ro \
		-e CANDIDATE_NAME="docker_test" \
		-e LOG_LEVEL=debug \
		--memory="512m" \
		--cpus="0.5" \
		--security-opt=no-new-privileges:true \
		--name echobox-dev \
		echobox:latest
	@echo "$(COLOR_YELLOW)URL: http://localhost:$(DOCKER_PORT)$(COLOR_RESET)"

.PHONY: docker-run-prod
docker-run-prod: docker-build-prod ## Build and run production container with strict security
	@echo "$(COLOR_BLUE)Starting production container...$(COLOR_RESET)"
	@mkdir -p sessions tasks
	@docker run -d \
		-p $(DOCKER_PORT):8080 \
		-v $(PWD)/sessions:/output \
		-v $(PWD)/tasks:/tasks:ro \
		-e CANDIDATE_NAME="${CANDIDATE_NAME:-prod_candidate}" \
		-e SESSION_TIMEOUT=7200 \
		--memory="512m" \
		--memory-reservation="256m" \
		--cpus="0.5" \
		--security-opt=no-new-privileges:true \
		--cap-drop=ALL \
		--cap-add=CHOWN \
		--cap-add=SETUID \
		--cap-add=SETGID \
		--restart=no \
		--name echobox-prod-$(shell date +%s) \
		echobox:prod
	@echo "$(COLOR_GREEN)✓ Production container started$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)Container: $(shell docker ps --format '{{.Names}}' -f name=echobox-prod | head -1)$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)URL: http://localhost:$(DOCKER_PORT)$(COLOR_RESET)"

.PHONY: docker-compose-up
docker-compose-up: ## Start with docker-compose (development)
	@echo "$(COLOR_BLUE)Starting with docker-compose...$(COLOR_RESET)"
	@mkdir -p sessions tasks
	@chmod 777 sessions 2>/dev/null || true
	@docker-compose up echobox-dev

.PHONY: docker-compose-prod
docker-compose-prod: ## Start with docker-compose (production)
	@echo "$(COLOR_BLUE)Starting production container with docker-compose...$(COLOR_RESET)"
	@mkdir -p sessions tasks
	@chmod 777 sessions 2>/dev/null || true
	@docker-compose up -d echobox-prod
	@echo "$(COLOR_GREEN)✓ Production container started$(COLOR_RESET)"
	@echo "$(COLOR_YELLOW)URL: http://localhost:${DOCKER_PORT:-8080}$(COLOR_RESET)"

.PHONY: docker-stop
docker-stop: ## Stop all echobox containers
	@echo "$(COLOR_BLUE)Stopping containers...$(COLOR_RESET)"
	@docker ps -q -f name=echobox | xargs -r docker stop || true
	@echo "$(COLOR_GREEN)✓ Containers stopped$(COLOR_RESET)"

.PHONY: docker-logs
docker-logs: ## Show container logs
	@docker logs -f echobox-dev 2>/dev/null || docker logs -f echobox-prod 2>/dev/null || echo "$(COLOR_YELLOW)No running containers$(COLOR_RESET)"

.PHONY: docker-exec
docker-exec: ## Execute shell in running container
	@docker exec -it echobox-dev /bin/bash 2>/dev/null || docker exec -it echobox-prod /bin/bash 2>/dev/null || echo "$(COLOR_YELLOW)No running containers$(COLOR_RESET)"

.PHONY: docker-clean
docker-clean: ## Remove echobox images and containers
	@echo "$(COLOR_BLUE)Cleaning Docker resources...$(COLOR_RESET)"
	@docker ps -a -q -f name=echobox | xargs -r docker rm -f || true
	@docker images echobox -q | xargs -r docker rmi -f || true
	@echo "$(COLOR_GREEN)✓ Docker resources cleaned$(COLOR_RESET)"

.PHONY: docker-inspect
docker-inspect: ## Inspect running container
	@docker inspect echobox-dev 2>/dev/null || docker inspect echobox-prod 2>/dev/null || echo "$(COLOR_YELLOW)No running containers$(COLOR_RESET)"

##@ Cleanup

.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(COLOR_BLUE)Cleaning...$(COLOR_RESET)"
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -f *.log
	@echo "$(COLOR_GREEN)✓ Clean complete$(COLOR_RESET)"

.PHONY: clean-all
clean-all: clean ## Clean everything including sessions
	@echo "$(COLOR_BLUE)Cleaning all artifacts including sessions...$(COLOR_RESET)"
	@rm -rf sessions/
	@rm -rf output/
	@echo "$(COLOR_GREEN)✓ All artifacts cleaned$(COLOR_RESET)"

##@ Utilities

.PHONY: health
health: ## Check server health
	@echo "$(COLOR_BLUE)Checking server health...$(COLOR_RESET)"
	@curl -s http://localhost:8080/health | jq '.' || echo "$(COLOR_YELLOW)⚠ Server not running or jq not installed$(COLOR_RESET)"

.PHONY: install
install: build ## Install binary to $GOPATH/bin
	@echo "$(COLOR_BLUE)Installing $(BINARY_NAME) to $(GOPATH)/bin...$(COLOR_RESET)"
	@cp $(BINARY_NAME) $(GOPATH)/bin/
	@echo "$(COLOR_GREEN)✓ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(COLOR_RESET)"

.PHONY: uninstall
uninstall: ## Uninstall binary from $GOPATH/bin
	@echo "$(COLOR_BLUE)Uninstalling $(BINARY_NAME)...$(COLOR_RESET)"
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(COLOR_GREEN)✓ Uninstalled$(COLOR_RESET)"

.PHONY: version
version: ## Show Go version
	@$(GO) version

.PHONY: info
info: ## Show project information
	@echo "$(COLOR_BOLD)Project Information$(COLOR_RESET)"
	@echo "  Binary name:  $(BINARY_NAME)"
	@echo "  Go version:   $(shell $(GO) version)"
	@echo "  Packages:     $(words $(PKG_LIST))"
	@echo "  Build dir:    $(BUILD_DIR)"

##@ Development Tools

.PHONY: dev-setup
dev-setup: ## Setup development environment
	@echo "$(COLOR_BLUE)Setting up development environment...$(COLOR_RESET)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)Installing golangci-lint...$(COLOR_RESET)"; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v entr >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)Install entr for watch support: brew install entr$(COLOR_RESET)"; \
	fi
	@if ! command -v jq >/dev/null 2>&1; then \
		echo "$(COLOR_YELLOW)Install jq for JSON parsing: brew install jq$(COLOR_RESET)"; \
	fi
	@$(GO) mod download
	@echo "$(COLOR_GREEN)✓ Development environment ready$(COLOR_RESET)"

.PHONY: pre-commit
pre-commit: fmt vet test ## Run pre-commit checks (fmt, vet, test)
	@echo "$(COLOR_GREEN)✓ Pre-commit checks passed$(COLOR_RESET)"
