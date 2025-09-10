---
name: Provider Compatibility Issue
about: Report compatibility issues with Terraform versions or other providers
title: "[COMPATIBILITY] - "
labels: compatibility, bug
assignees: ''

---

## Provider Compatibility Issue

### Description

A clear and concise description of the compatibility issue.

### Environment

- **Terraform version**: [e.g. v1.5.0]
- **Provider version**: [e.g. v1.0.0]
- **Wallix Bastion version**: [e.g. v3.12.0]
- **Operating System**: [e.g. Ubuntu 22.04, Windows 11, macOS 13]
- **Go version** (if building from source): [e.g. 1.21.0]

### Other providers (if relevant)

List any other providers being used that might conflict:

- Provider name: version
- Provider name: version

### Configuration

```hcl
# Paste your Terraform configuration here
terraform {
  required_providers {
    wallix-bastion = {
      # your configuration
    }
  }
}
```

### Error Output

```text
Paste the complete error message here
```

### Expected Behavior

What should happen instead?

### Additional Context

Any additional information that might help resolve the compatibility issue.

### Workaround

If you found a workaround, please describe it here.
