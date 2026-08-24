# Define variables for maintainability

PROVIDER_NAME := wallix-bastion

# Version dynamique basée sur git

VERSION := $(shell git describe --tags --exact-match 2>/dev/null || echo "dev")

# Ou version avec branche pour dev

DEV_VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")

# Version conditionnelle

ifeq ($(ENV),production)
    PROVIDER_VERSION := 1.0.0
else
    PROVIDER_VERSION := 0.0.0-dev
endif

# The go build output binary name

BINARY_NAME := terraform-provider-$(PROVIDER_NAME)

# Use a -ldflags string for versioning during compilation

LDFLAGS_STRING := "-X main.version=$(VERSION)"

.PHONY: build install test testacc test-coverage fmt lint lint-fix lint-markdown lint-security vet clean setup-dev setup-check docs docs-verify build-all test-all maintenance prepare-release release-patch release-minor release-major release ci-check dev-check

# Default target

default: build

# Build the provider

# This target now includes version information from git via LDFLAGS

build:
	go build -ldflags=$(LDFLAGS_STRING) -o $(BINARY_NAME)

# Build for all platforms

build-all:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags=$(LDFLAGS_STRING) -o dist/$(BINARY_NAME)_darwin_amd64
	GOOS=darwin GOARCH=arm64 go build -ldflags=$(LDFLAGS_STRING) -o dist/$(BINARY_NAME)_darwin_arm64
	GOOS=linux GOARCH=amd64 go build -ldflags=$(LDFLAGS_STRING) -o dist/$(BINARY_NAME)_linux_amd64
	GOOS=linux GOARCH=arm64 go build -ldflags=$(LDFLAGS_STRING) -o dist/$(BINARY_NAME)_linux_arm64
	GOOS=windows GOARCH=amd64 go build -ldflags=$(LDFLAGS_STRING) -o dist/$(BINARY_NAME)_windows_amd64.exe

# Install locally for development

# This target is now cross-platform, automatically detecting your OS and architecture

install: build

# Use the Go environment variables directly to create the correct plugin path

	mkdir -p  ~/.terraform.d/plugins/terraform.local/local/$(PROVIDER_NAME)/$(PROVIDER_VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)

# Copy the binary to the correct location

	cp $(BINARY_NAME) ~/.terraform.d/plugins/terraform.local/local/$(PROVIDER_NAME)/$(PROVIDER_VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

# Run unit tests

test:
	@if command -v go-test-report >/dev/null 2>&1; then \
	  go test -v ./... -json | go-test-report; \
	else \
	  echo "go-test-report not found, running tests without report generation"; \
	  go test -v ./...; \
	fi

# Run tests with coverage

test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run acceptance tests

testacc:
	TF_ACC=1 WALLIX_INSECURE_SKIP_VERIFY=true go test -v ./bastion -timeout 120m

# Run all tests

test-all: test testacc

# Format code

fmt:
	go fmt ./...
	terraform fmt -recursive examples/

# Run linters with configuration matching CI

lint:
	golangci-lint run --config .golangci.yml --verbose --timeout 5m

# Run linters and auto-fix issues when possible

lint-fix:
	golangci-lint run --config .golangci.yml --fix --timeout 5m

# Lint markdown files

lint-markdown:
	@if command -v markdownlint >/dev/null 2>&1; then \
	  markdownlint **/*.md --ignore node_modules --ignore .terraform --ignore WIP --config .markdownlint.json; \
	else \
	  echo "Warning: markdownlint not found. Install with: npm install -g markdownlint-cli"; \
	  exit 1; \
	fi

# Run security vulnerability checks

lint-security:
	@echo "Running vulnerability checks..."
	@if command -v govulncheck >/dev/null 2>&1; then \
	  govulncheck ./...; \
	else \
	  echo "Warning: govulncheck not found. Installing..."; \
	  go install golang.org/x/vuln/cmd/govulncheck@latest; \
	  govulncheck ./...; \
	fi

# Run go vet

vet:
	go vet ./...

# Clean build artifacts

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
	rm -f coverage.out coverage.html

# Setup development environment

setup-dev:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest
	@echo ""
	@echo "✓ Go tools installed successfully"
	@echo ""
	@echo "Optional tools (install manually):"
	@echo "  - Terraform CLI: https://www.terraform.io/downloads"
	@echo "  - markdownlint: npm install -g markdownlint-cli"
	@echo "  - markdownlint-cli2: npm install -g markdownlint-cli2"
	@echo ""
	@echo "Verify installation:"
	@echo "  golangci-lint version: $$(golangci-lint --version 2>/dev/null || echo 'not found')"
	@echo "  govulncheck: $$(govulncheck -version 2>/dev/null || echo 'not found')"
	@echo "  tfplugindocs: $$(tfplugindocs --version 2>/dev/null || echo 'installed')"

# Verify development environment setup

setup-check:
	@echo "Checking development environment..."
	@echo ""
	@echo "Required Go tools:"
	@command -v golangci-lint >/dev/null 2>&1 && echo "  ✓ golangci-lint: $$(golangci-lint --version | head -1)" || echo "  ✗ golangci-lint: not found"
	@command -v govulncheck >/dev/null 2>&1 && echo "  ✓ govulncheck: installed" || echo "  ✗ govulncheck: not found"
	@command -v tfplugindocs >/dev/null 2>&1 && echo "  ✓ tfplugindocs: installed" || echo "  ✗ tfplugindocs: not found"
	@echo ""
	@echo "Optional tools:"
	@command -v terraform >/dev/null 2>&1 && echo "  ✓ terraform: installed" || echo "  ✗ terraform: not found"
	@command -v markdownlint >/dev/null 2>&1 && echo "  ✓ markdownlint: installed" || echo "  ✗ markdownlint: not found (npm install -g markdownlint-cli)"
	@echo ""
	@command -v golangci-lint >/dev/null 2>&1 && command -v govulncheck >/dev/null 2>&1 && command -v tfplugindocs >/dev/null 2>&1 && echo "✓ All required tools installed!" || echo "⚠ Some required tools are missing. Run 'make setup-dev' to install them."

# Generate documentation

docs:
	@echo "Generating documentation with tfplugindocs..."
	@if command -v tfplugindocs >/dev/null 2>&1; then \
	  tfplugindocs generate; \
	else \
	  echo "Error: tfplugindocs not found. Install with: go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest"; \
	  exit 1; \
	fi

# Verify documentation quality

docs-verify: docs
	@echo "Verifying documentation quality..."
	@if command -v markdownlint >/dev/null 2>&1; then \
	  markdownlint docs/ --fix; \
	  echo "Documentation linting completed"; \
	else \
	  echo "Warning: markdownlint not found. Install with: npm install -g markdownlint-cli"; \
	  echo "Skipping markdown linting..."; \
	fi
	@echo "Checking for template completeness..."
	@tfplugindocs validate
	@echo "Documentation verification completed"

# Analyze API coverage

coverage-api:
	@echo "Analyzing API coverage..."
	@if [ ! -f "tools/coverage-analyzer/main.go" ]; then \
	  echo "Coverage analyzer not found. Please create it first."; \
	  exit 1; \
	fi
	@cd tools/coverage-analyzer && go run main.go -provider ../../ -verbose

# Setup coverage analysis tools

setup-coverage:
	@mkdir -p tools/coverage-analyzer
	@echo "Coverage analyzer directory created"
	@echo "Please add the coverage analyzer code to tools/coverage-analyzer/main.go"

# Maintenance tasks

maintenance:
	@./scripts/maintenance.sh

# Release preparation scripts

prepare-release:
	@./scripts/prepare-release.sh --dry-run

release-patch:
	@./scripts/prepare-release.sh --patch

release-minor:
	@./scripts/prepare-release.sh --minor

release-major:
	@./scripts/prepare-release.sh --major

# For a specific version, e.g. `make release ARGS="--version v0.15.0"`.
# Note: `make release-minor --version vX` does NOT work - `--version` is parsed
# by make itself (it prints make's own version and exits), never reaching the
# script. Flags must be passed through the ARGS variable instead.
release:
	@./scripts/prepare-release.sh $(ARGS)

# Quick development checks

dev-check: lint test build
	@echo "Development checks completed successfully"

# Run all quality checks (like CI)

ci-check: lint lint-security test
	@echo "CI-style checks completed successfully"

# Update dependencies

update-deps:
	@./scripts/maintenance.sh deps
