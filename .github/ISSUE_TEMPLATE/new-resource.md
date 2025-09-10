---
name: New Resource/Data Source Request
about: Request a new resource or data source for the provider
title: "[NEW RESOURCE] - "
labels: enhancement, new-resource
assignees: ''

---

## New Resource/Data Source Request

### Type of Request

- [ ] New Resource (terraform resource "wallix-bastion_...")
- [ ] New Data Source (terraform data "wallix-bastion_...")
- [ ] Enhancement to existing resource/data source

### Wallix Bastion Feature

**What Wallix Bastion feature/API should this resource manage?**

Describe the specific Wallix Bastion functionality this would expose.

### API Endpoints

**Which Wallix Bastion API endpoints would be used?**

- `GET /api/endpoint` - for reading
- `POST /api/endpoint` - for creation
- `PUT /api/endpoint` - for updates
- `DELETE /api/endpoint` - for deletion

### Proposed Resource/Data Source Name

```hcl
resource "wallix-bastion_your_resource_name" "example" {
  # proposed schema
}

# or

data "wallix-bastion_your_data_source_name" "example" {
  # proposed schema
}
```

### Proposed Schema

**What attributes should this resource/data source have?**

```hcl
# Example schema
resource "wallix-bastion_example" "test" {
  name        = "example"
  description = "An example resource"
  enabled     = true
  
  # Computed attributes
  id           = "computed"
  created_date = "computed"
}
```

### Use Case

**How would this resource/data source be used?**

Describe the use case and provide examples of how users would use this in their Terraform configurations.

### Priority

- [ ] Critical - Blocks adoption
- [ ] High - Needed for common use cases
- [ ] Medium - Would be nice to have
- [ ] Low - Future enhancement

### Additional Context

Any other information that would help implement this resource/data source.

### Related Issues

Link to any related issues or feature requests.
