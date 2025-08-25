# Define variables for maintainability
PROVIDER_NAME := wallix-bastion
PROVIDER_VERSION := 1.0.0
# The go build output binary name
BINARY_NAME := terraform-provider-$(PROVIDER_NAME)

# Get the version from git, defaulting to "dev"
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Use a -ldflags string for versioning during compilation
LDFLAGS_STRING := "-X main.version=$(VERSION)"

.PHONY: build install test testacc test-coverage fmt lint vet clean setup-dev docs build-all test-all

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
	mkdir -p ~/.terraform.d/plugins/local/$(PROVIDER_NAME)/$(PROVIDER_VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)
	# Copy the binary to the correct location
	cp $(BINARY_NAME) ~/.terraform.d/plugins/local/$(PROVIDER_NAME)/$(PROVIDER_VERSION)/$(shell go env GOOS)_$(shell go env GOARCH)/

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
	TF_ACC=1 go test -v ./bastion -timeout 120m

# Run all tests
test-all: test testacc

# Format code
fmt:
	go fmt ./...
	terraform fmt -recursive examples/

# Run linters
lint:
	golangci-lint run

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
	go mod download
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Generate documentation
docs:
	go generate ./...