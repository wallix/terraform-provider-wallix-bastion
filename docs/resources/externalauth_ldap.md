# wallix-bastion_externalauth_ldap Resource

Provides a LDAP externaulauth resource.

## Argument Reference

The following arguments are supported:

* `authentication_name` - (Required)(`String`) The authentication name.
* `cn_attribute` - (Required)(`String`) The username attribute.
* `host` - (Required)(`String`) The host name.
* `ldap_base` - (Required)(`String`) The LDAP base scheme.
* `login_attribute` - (Required)(`String`) The login attribute.
* `port` - (Required)(`Int`) The port number.
* `timeout` - (Required)(`Int`) LDAP timeout.
* `ca_certificate` - (Optional)(`String`) CA certificate.
* `certificate` - (Optional)(`String`) Client certificate.
* `description` - (Optional)(`String`) Description of the authentication.
* `is_active_directory` - (Optional)(`Bool`) This LDAP uses an active directory.
* `is_anonymous_access` - (Optional)(`Bool`) The user is anonymous.
* `is_protected_user` - (Optional)(`Bool`) The AD user is protected.
* `is_ssl` - (Optional)(`Bool`) This LDAP is secure (with SSL/TLS).
* `is_starttls` - (Optional)(`Bool`) This LDAP uses STARTTLS.
* `login` - (Optional)(`String`) The login. Required if is_anonymous_access = false.
* `password` - (Optional)(`String`) The password. Required if is_anonymous_access = false. **Value can't refresh**
* `private_key` - (Optional)(`String`) Client key.
* `use_primary_auth_domain` - (Optional)(`Bool`) Use the primary auth domain.

## Attribute Reference
* `id` - (`String`) Internal id of externalauth in bastion.

## Import

LDAP externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```
$ terraform import wallix-bastion_user.server1 server1
```
