---
name: Performance Issue
about: Report performance problems or optimization requests
title: "[PERFORMANCE] - "
labels: performance
assignees: ''

---

## Performance Issue

### Description

Describe the performance issue you're experiencing.

### Resource/Operation Affected

- [ ] Provider initialization
- [ ] Resource creation
- [ ] Resource updates
- [ ] Resource deletion
- [ ] Data source queries
- [ ] Bulk operations
- [ ] Other: ___________

### Performance Metrics

If you have specific measurements, please include them:

- **Operation duration**: [e.g. 5 minutes]
- **Number of resources**: [e.g. 100 users]
- **Memory usage**: [e.g. 2GB]
- **API calls made**: [e.g. 500 requests]

### Configuration

```hcl
# Paste relevant Terraform configuration
resource "wallix-bastion_user" "example" {
  # your configuration
}
```

### Environment

- **Provider version**: [e.g. v1.0.0]
- **Terraform version**: [e.g. v1.5.0]
- **Wallix Bastion version**: [e.g. v3.12.0]
- **Infrastructure size**: [e.g. 1000 users, 50 devices]

### Expected Performance

What performance did you expect?

### Actual Performance

What performance are you actually seeing?

### Logs/Debug Output

```text
# Enable TF_LOG=DEBUG and paste relevant output
# Remove any sensitive information
```

### Potential Optimizations

If you have ideas for how to improve performance, please share them here.
