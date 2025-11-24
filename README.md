# Terraform Provider for Wallix Bastion

![Wallix Logo](https://raw.githubusercontent.com/wallix/terraform-provider-wallix-bastion/refs/heads/main/assets/LOGO_WALLIX.png)

A Terraform provider for managing Wallix Bastion resources

[![Go Report Card](https://goreportcard.com/badge/github.com/wallix/terraform-provider-wallix-bastion)](https://goreportcard.com/report/github.com/wallix/terraform-provider-wallix-bastion)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4.svg)](https://registry.terraform.io/providers/wallix/wallix-bastion/latest)

## Overview

The Terraform Wallix Bastion provider allows you to manage Wallix Bastion resources such as users, groups, authorizations, and more through Infrastructure as Code.

## Requirements

- Visit the official [Terraform website](https://www.terraform.io/) for downloads
- [Go](https://golang.org/doc/install) `v1.22` or `v1.23` (for development)

### From Terraform Registry

```hcl
terraform {
  required_providers {
    wallix-bastion = {
      source  = "wallix/wallix-bastion"
      version = "~> 0.14.0"
    }
  }
}

provider "wallix-bastion" {
  ip          = "your-bastion-host"
  user        = "<user>"
  token       = "<your-api-token>"
}
```

### Local Development Installation

```bash
# Clone the repository
git clone https://github.com/wallix/terraform-provider-wallix-bastion.git
cd terraform-provider-wallix-bastion

```bash
# Build and install the provider locally
make install
```

> **Note:** When testing your locally built provider, you may need to explicitly specify the local source and version in your Terraform configuration.
> This ensures Terraform uses your development build instead of the published version.

**Example:**

```hcl
terraform {
  required_providers {
    wallix-bastion = {
      source  = "terraform.local/local/wallix-bastion"
      version = "0.0.0-dev"
    }
  }
}
```

## Building the Provider

### Prerequisites

For basic building and testing:

- **Go** 1.22 to 1.24
- **Make**
- **Git**

For full development environment (installed via `make setup-dev`):

- **golangci-lint** - Code linting
- **govulncheck** - Security vulnerability scanning
- **tfplugindocs** - Documentation generation

Optional tools:

- **Terraform CLI** - For manual testing
- **markdownlint** - Markdown linting (`npm install -g markdownlint-cli`)
- **go-test-report** - Enhanced test output

### Quick Build

```bash
# Build the provider
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Development Build

For manual development without using Makefile:

```bash
# Install dependencies
go mod download

# Format code
go fmt ./...

# Run linters
golangci-lint run --config .golangci.yml

# Build with version info
go build -ldflags="-X main.version=dev" -o terraform-provider-wallix-bastion
```

**Recommended:** Use `make` commands instead for a consistent development experience.

## Testing

### Running Unit Tests

```bash
# Run all unit tests
make test

# Run tests with HTML coverage report
make test-coverage

# Run specific test
go test -v ./bastion -run TestAccResourceAuthorization_basic

# Run all tests (unit + acceptance)
make test-all
```

**Note:** If `go-test-report` is installed, `make test` will generate a formatted test report.

### Running Acceptance Tests

Acceptance tests require a running Wallix Bastion instance.

```bash
# Set environment variables
export WALLIX_BASTION_HOST="your-bastion-host"
export WALLIX_BASTION_USER="admin"
export WALLIX_BASTION_TOKEN="<your-api-token>"
export WALLIX_BASTION_API_VERSION="v3.12"

# Run acceptance tests
make testacc

# Run specific acceptance test
TF_ACC=1 go test -v ./bastion -run TestAccResourceAuthorization_sessionSharing
```

### Test Environment Setup

1. **Create environment file from template:**

   ```bash
   # Copy the example environment file
   cp .env.test.example .env.test
   
   # Edit with your Bastion credentials
   vim .env.test  # or use your preferred editor
   ```

2. **Load environment variables:**

   ```bash
   # Source the environment file
   source .env.test
   
   # Verify configuration
   echo "Testing against: $WALLIX_BASTION_HOST"
   echo "API Version: $WALLIX_BASTION_API_VERSION"
   ```

3. **Verify Bastion connectivity:**

   ```bash
   # Test API endpoint
   curl -k https://$WALLIX_BASTION_HOST/api/version
   ```

4. **Run tests:**

   ```bash
   # Run unit tests
   make test
   
   # Run acceptance tests (requires configured .env.test)
   make testacc
   
   # Run all tests
   make test-all
   ```

**Security Note:** Never commit `.env.test` to version control. It's already in `.gitignore`.

## Local Development Workflow

### 1. Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/wallix/terraform-provider-wallix-bastion.git
cd terraform-provider-wallix-bastion

# Install dependencies and development tools
make setup-dev

# Verify installation
make setup-check
```

### 2. Make Changes

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ...

# Format and lint
make fmt
make lint

# Run tests
make test

# Run security checks
make lint-security
```

### 3. Test Locally

```bash
# Build and install locally
make install

# Test with your Terraform configuration
cd examples/

# Choose an example directory, e.g., authorization
cd authorization

# Update the provider to use the development build

# terraform {
#   required_version = ">= 1.0"
#   required_providers {
#     wallix-bastion = {
#       # source  = "wallix/wallix-bastion"
#       # version = "0.14.7"
#       source  = "terraform.local/local/wallix-bastion"
#       version = "0.0.0-dev"
#     }
#   }
# }

terraform init
terraform plan
terraform apply
```

### 4. Submit Changes

```bash
# Run full checks (like CI)
make ci-check

# Or run all tests
make test-all

# Commit changes
git add .
git commit -m "feat: your feature description"
git push origin feature/your-feature-name

# Create pull request
```

## Makefile Commands

### Build Commands

```bash
make build          # Build the provider with version info
make build-all      # Build for all platforms (darwin/linux/windows, amd64/arm64)
make install        # Install the provider locally for development
make clean          # Clean build artifacts and coverage files
```

### Code Quality Commands

```bash
make fmt            # Format Go code and Terraform examples
make lint           # Run golangci-lint with CI configuration
make lint-fix       # Run golangci-lint and auto-fix issues
make lint-markdown  # Lint markdown files (requires markdownlint)
make lint-security  # Run vulnerability checks with govulncheck
make vet            # Run go vet
```

### Test Commands

```bash
make test           # Run unit tests (with go-test-report if available)
make test-coverage  # Run tests with coverage report (HTML output)
make testacc        # Run acceptance tests (requires Bastion instance)
make test-all       # Run all tests (unit + acceptance)
```

### Development Commands

```bash
make setup-dev      # Setup development environment (install tools)
make setup-check    # Verify development environment setup
make dev-check      # Quick development checks (lint + test + build)
make ci-check       # Run CI-style checks (lint + security + test)
```

### Documentation Commands

```bash
make docs           # Generate documentation with tfplugindocs
make docs-verify    # Verify and lint documentation quality
```

### Maintenance and Release Commands

```bash
make maintenance    # Run maintenance tasks (deps + lint + test + build)
make update-deps    # Update Go dependencies only
make prepare-release # Dry-run release preparation
make release-patch  # Prepare patch release (X.Y.Z+1)
make release-minor  # Prepare minor release (X.Y+1.0)
make release-major  # Prepare major release (X+1.0.0)
```

**Note:** See [RELEASE.md](./RELEASE.md) for detailed release process documentation.

## Documentation

- [Provider Documentation](https://registry.terraform.io/providers/wallix/wallix-bastion/latest/docs)
- [Documentation Generation Guide](./DOCUMENTATION.md) - How to generate and verify documentation
- [API Documentation](https://docs.wallix.com/)
- [Examples](./examples/)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Quick Start for Contributors

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Run quality checks (`make dev-check` or `make ci-check`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Development Tools Setup

```bash
# Install all development tools
make setup-dev

# Verify your setup
make setup-check
```

## Version Compatibility

| Provider Version | Terraform Version | Go Version | Wallix Bastion API |
|------------------|-------------------|------------|-------------------|
| >= 0.14.0        | >= 1.0           | 1.22-1.24  | v3.8, v3.12       |
| 0.13.x           | >= 0.14          | 1.19-1.21  | v3.3, v3.6        |

## License

This project is licensed under the Mozilla Public License 2.0 - see the [LICENSE](LICENSE) file for details.

## Special Thanks

We would like to greatly thanks:

- [Claranet](https://www.claranet.com/) for their great work on this provider!
- The Terraform community for their continuous support and contributions

## Support

- 📖 [Documentation](https://registry.terraform.io/providers/wallix/wallix-bastion/latest/docs)
- 🐛 [Issue Tracker](https://github.com/wallix/terraform-provider-wallix-bastion/issues)
- 💬 [Discussions](https://github.com/wallix/terraform-provider-wallix-bastion/discussions)
