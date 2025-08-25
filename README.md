# Terraform Provider for Wallix Bastion

![Wallix Logo](https://raw.githubusercontent.com/wallix/terraform-provider-wallix-bastion/refs/heads/main/assets/LOGO_WALLIX_2024_black%2Borange.png)

A Terraform provider for managing Wallix Bastion resources

[![Go Report Card](https://goreportcard.com/badge/github.com/wallix/terraform-provider-wallix-bastion)](https://goreportcard.com/report/github.com/wallix/terraform-provider-wallix-bastion)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)
[![Terraform Registry](https://img.shields.io/badge/terraform-registry-623CE4.svg)](https://registry.terraform.io/providers/wallix/wallix-bastion/latest)

## Overview

The Terraform Wallix Bastion provider allows you to manage Wallix Bastion resources such as users, groups, authorizations, and more through Infrastructure as Code.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) `v1.22` or `v1.23` (for development)
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

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

# Build and install locally
make install
```

## Building the Provider

### Prerequisites

Ensure you have the following installed:

- Go 1.22 or 1.23
- Make
- Git

### Build Commands

```bash
# Build the provider
make build

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Development Build

```bash
# Install development dependencies
go mod download

# Format code
make fmt

# Run linters
make lint

# Build development version
go build -o terraform-provider-wallix-bastion
```

## Testing

### Running Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test
go test -v ./bastion -run TestAccResourceAuthorization_basic
```

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

1. **Set up test environment variables:**

   ```bash
   export WALLIX_BASTION_HOST="your-test-bastion"
   export WALLIX_BASTION_TOKEN="<your-test-token>"
   export WALLIX_BASTION_USER="admin"
   export WALLIX_BASTION_API_VERSION="v3.8"
   export TF_ACC=1
   ```

2. **Create test configuration:**

   ```bash
   # Copy example configuration
   cp examples/authorization_test.tf test.tf

   # Edit with your test values
   vim test.tf
   ```

3. **Run manual tests:**

   ```bash
   terraform init
   terraform plan
   terraform apply
   terraform destroy
   ```

## Local Development Workflow

### 1. Setup Development Environment

```bash
# Clone and setup
git clone https://github.com/wallix/terraform-provider-wallix-bastion.git
cd terraform-provider-wallix-bastion

# Install dependencies
go mod download
go mod tidy

# Setup pre-commit hooks (optional)
make setup-dev
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
```

### 3. Test Locally

```bash
# Build and install locally
make install

# Test with your Terraform configuration
cd examples/

# Choose an example directory, e.g., authorization
cd authorization

terraform init
terraform plan
terraform apply
```

### 4. Submit Changes

```bash
# Run full test suite
make test-all

# Commit changes
git add .
git commit -m "feat: your feature description"
git push origin feature/your-feature-name

# Create pull request
```

## Makefile Commands

```bash
# Build commands
make build          # Build the provider
make build-all      # Build for all platforms

# Quality commands
make fmt            # Format Go code
make lint           # Run linters
make vet            # Run go vet

# Test commands
make test           # Run unit tests
make test-coverage  # Run tests with coverage
make testacc        # Run acceptance tests
make test-all       # Run all tests

# Development commands
make clean          # Clean build artifacts
make setup-dev      # Setup development environment
make install        # Install the provider locally
make docs           # Generate documentation
```

## Documentation

- [Provider Documentation](https://registry.terraform.io/providers/wallix/wallix-bastion/latest/docs)
- [API Documentation](https://docs.wallix.com/)
- [Examples](./examples/)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes and add tests
4. Run the test suite (`make test-all`)
5. Commit your changes (`git commit -m 'Add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Version Compatibility

| Provider Version | Terraform Version | Go Version | Wallix Bastion API |
|------------------|-------------------|------------|-------------------|
| >= 0.14.0        | >= 1.0           | 1.22-1.23  | v3.8, v3.12      |
| 0.13.x           | >= 0.14          | 1.19-1.21  | v3.3, v3.6       |

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
