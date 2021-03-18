# wallix-bastion_externalauth_radius Resource

Provides a RADIUS externaulauth resource.

## Argument Reference

The following arguments are supported:

* `authentication_name` - (Required)(`String`) The authentication name.
* `host` - (Required)(`String`) The host name.
* `port` - (Required)(`Int`) The port number.
* `secret` - (Optional)(`String`) The secret.
* `timeout` - (Required)(`Int`) RADIUS timeout.
* `description` - (Optional)(`String`) Description of the authentication.
* `use_primary_auth_domain` - (Optional)(`Bool`) Use the primary auth domain.

## Attribute Reference

* `id` - (`String`) Internal id of externalauth in bastion.

## Import

RADIUS externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```
$ terraform import wallix-bastion_externalauth_radius.server1 server1
```
