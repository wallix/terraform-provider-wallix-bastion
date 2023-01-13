# wallix-bastion_version Data Source

Get information on Wallix version.

## Example Usage

```hcl
data "wallix-bastion_version" "version" {}
```

## Attribute Reference

- **version** (String)  
  The REST API version.
- **version_decimal** (String)  
  The REST API version as decimal number.
- **wab_version** (String)  
  The WALLIX Bastion version (format: X.Y).
- **wab_version_decimal** (String)  
  The WALLIX Bastion version as decimal number.
- **wab_version_hotfix** (String)  
  The WALLIX Bastion version with hotfix level (format: X.Y.Z, Z being the hotfix level).
- **wab_version_hotfix_decimal** (String)  
  The WALLIX Bastion version with hotfix level as decimal.
- **wab_complete_version** (String)  
  The WALLIX Bastion complete version, with hotfix level and build date.
