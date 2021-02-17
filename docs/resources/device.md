# wallix-bastion_device Resource

Provides a device resource.

## Argument Reference

The following arguments are supported:

* `device_name` - (Required)(`String`) The device name.
* `host` - (Required)(`String`) The device host address.
* `alias` - (Optional)(`String`) The device alias.
* `description` - (Optional)(`String`) The device description.

## Attribute Reference

* `id` - (`String`) Internal id of device in bastion.
* `local_domains` - (`ListOfNestedBlock`) List of localdomain
  * `id` - (`String`) Internal id of local domain in bastion.
  * `domain_name` - (`String`) The domain name.
  * `admin_account` - (`String`) The administrator account used to change passwords on this domain (format: "account_name@domain_name").
  * `ca_public_key` - (`String`) The ssh public key of the signing authority for the ssh keys for accounts in the domain.
  * `description` - (`String`) The domain description.
  * `enable_password_change` - (`Bool`) Enable the change of password on this domain.
  * `password_change_policy` - (`String`) The name of password change policy for this domain.
  * `password_change_plugin` - (`String`) The name of plugin used to change passwords on this domain.
  * `password_change_plugin_parameters` - (`NestedBlock`) Parameters for the plugin used to change credentials.
* `services` - (`ListOfNestedBlock`) List of service
  * `id` - (`String`) Internal id of service in bastion.
  * `service_name` - (`String`) The service name.
  * `connection_policy` - (`String`) The connection policy name.
  * `port` - (`Int`) The port number.
  * `protocol` - (`String`) The protocol.
  * `global_domains` - (`ListOfString`) The global domains names.
  * `subprotocols` - (`ListOfString`) The sub protocols for 'SSH', 'RDP' protocol.

## Import

Device can be imported using an id made up of `<device_name>`, e.g.

```
$ terraform import wallix-bastion_device.server1 server1
```
