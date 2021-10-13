# wallix-bastion_ldapmapping Resource

Provides a ldapmapping resource.

## Example Usage

```hcl
# Configure a ldapmapping
resource wallix-bastion_ldapmapping "test" {
  domain     = "domain.local"
  user_group = "group1"
  ldap_group = "CN=Test,OU=Group,DC=domain,DC=local"
}
```

## Argument Reference

The following arguments are supported:

- **domain** (Required, String, Forces new resource)  
  The name of the domain for which the mapping is defined.
- **user_group** (Required, String, Forces new resource)  
  The name of the Bastion users group.
- **ldap_group** (Required, String, Forces new resource)  
  The name (distinguished name - DN) of the LDAP group, `*` means fallback mapping.

## Attribute Reference

- **id** (String)  
  An identifier for the resource with format `<domain>/<user_group>/<ldap_group>`

## Import

ldapmapping can be imported using an id made up of `<domain>/<user_group>/<ldap_group>`, e.g.

```shell
terraform import wallix-bastion_ldapmapping.test 'domain.local/group1/CN=test,OU=example,DC=com'
```
