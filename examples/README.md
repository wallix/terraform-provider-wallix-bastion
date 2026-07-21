# Terraform Provider WALLIX Bastion Examples

This directory contains examples that demonstrate how to use the WALLIX Bastion Terraform provider in various scenarios.

## Examples Overview

| Example | Description | Complexity |
|---------|-------------|------------|
| [basic](./basic/) | Basic setup with minimal configuration | Beginner |
| [authorization](./authorization/) | User and target groups with authorizations | Intermediate |
| [session-sharing](./session-sharing/) | Session sharing functionality | Intermediate |
| [approval-workflow](./approval-workflow/) | Authorization with approval workflow | Advanced |
| [complete-setup](./complete-setup/) | Comprehensive production-ready setup | Advanced |

## Prerequisites

1. **WALLIX Bastion instance** running and accessible
2. **API credentials** with appropriate permissions
3. **Terraform** >= 1.0 installed

## Quick Start

1. **Choose an example** based on your needs
2. **Copy the example** to your working directory:

   ```bash
   cp -r examples/basic/ my-bastion-config/
   cd my-bastion-config/
   ```

3. **Configure variables** by creating `terraform.tfvars`:

   ```hcl
   bastion_ip    = "192.168.1.100"
   bastion_token = "your-api-token"
   ```

4. **Deploy**:

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Common Variables

Most examples use these common variables:

- `bastion_ip` - IP address or hostname of your WALLIX Bastion
- `bastion_token` - API token for authentication
- `api_version` - API version (default: "v3.12")

## Environment Variables

You can also set these environment variables instead of using `terraform.tfvars`:

```bash
export TF_VAR_bastion_ip="192.168.1.100"
export TF_VAR_bastion_token="your-api-token"
export TF_VAR_api_version="v3.12"
```

## Testing Examples

Each example includes basic testing instructions. For comprehensive testing:

```bash
# Validate configuration
terraform validate

# Check what will be created
terraform plan

# Apply changes
terraform apply

# Clean up
terraform destroy
```

## Contributing

When adding new examples:

1. Follow the directory structure pattern
2. Include comprehensive README.md
3. Use variables for all configurable values
4. Add appropriate outputs
5. Test thoroughly before submitting
