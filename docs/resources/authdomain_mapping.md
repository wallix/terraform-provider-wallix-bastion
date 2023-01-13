# wallix-bastion_authdomain_mapping Resource

Provides a auth domain mapping resource.

## Example Usage

```hcl
# Configure a auth domain mapping
resource "wallix-bastion_authdomain_mapping" "test" {
  domain_id      = "xxxxxxxx"
  user_group     = "group1"
  external_group = "CN=Test,OU=Group,DC=domain,DC=local"
}
```

## Argument Reference

The following arguments are supported:

- **domain_id** (Required, String, Forces new resource)  
  ID of auth domain.
- **user_group** (Required, String)  
  The name of the Bastion users group.
- **external_group** (Required, String)  
  The name of the external group (LDAP/AD: Distinguished Name, Azure AD: name or ID),
  "*" means fallback mapping.

## Attribute Reference

- **id** (String)  
  Internal id of auth domain mapping in bastion.
- **domain** (String)  
  The name of the domain for which the mapping is defined.

## Import

Auth domain mapping can be imported using an id made up of `<domain_id>/<user_group>`, e.g.

```shell
terraform import wallix-bastion_authdomain_mapping.test 'xxxxxxxx/group1'
```
