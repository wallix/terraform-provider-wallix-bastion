# wallix-bastion_connection_policy Resource

Provides a connection_policy resource.

## Argument Reference

The following arguments are supported:

* `connection_policy_name` - (Required)(`String`) The connection policy name.
* `description` - (Optional)(`String`) The connection policy description.
* `protocol` - (Optional)(`String`) The connection policy protocol.
* `authentication_methods` - (Optional)(`ListOfString`) The allowed authentication methods.
* `options` - (Optional)(`String`) Options for the connection policy. Need to be a valid JSON.

## Attribute Reference

* `id` - (`String`) Internal id of connection policy in bastion.

## Import

Connection policy can be imported using an id made up of `<connection_policy_name>`, e.g.

```
$ terraform import wallix-bastion_connection_policy.pol example
```
