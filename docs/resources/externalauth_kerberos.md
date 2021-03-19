# wallix-bastion_externalauth_kerberos Resource

Provides a Kerberos externaulauth resource.

## Argument Reference

The following arguments are supported:

* `authentication_name` - (Required)(`String`) The authentication name.
* `host` - (Required)(`String`) The host name.
* `ker_dom_controller` - (Required)(`String`) Kerberos domain controller whose role is torecognizes the tickets issued bythe Key Distribution Center.
* `port` - (Required)(`Int`) The port number.
* `kerberos_password` - (Optional, Force new resource)(`Bool`) Use KERBEROS-PASSWORD protocol.
* `description` - (Optional)(`String`) Description of the authentication.
* `login_attribute` - (Optional)(`String`) The login attribute.
* `use_primary_auth_domain` - (Optional)(`Bool`) Use the primary auth domain.

## Attribute Reference

* `id` - (`String`) Internal id of externalauth in bastion.

## Import

Kerberos externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```
$ terraform import wallix-bastion_externalauth_kerberos.server1 server1
```
