# wallix-bastion_ldapmapping Resource

Provides a ldapmapping resource.

## Argument Reference

The following arguments are supported:

* `domain` - (Required, Forces new resource)(`String`) The name of the domain for which the mapping is defined.
* `user_group` - (Required, Forces new resource)(`String`) The name of the Bastion users group.
* `ldap_group` - (Required, Forces new resource)(`String`) The name (distinguished name - DN) of the LDAP group, "*" means fallback mapping.


## Attribute Reference

* `id` - (`String`) = `<domain>/<user_group>/<ldap_group>`

## Import

ldapmapping can be imported using an id made up of `<domain>/<user_group>/<ldap_group>`, e.g.

```
$ terraform import wallix-bastion_ldapmapping.test 'example.com/group1/CN=test,OU=example,DC=com'
```
