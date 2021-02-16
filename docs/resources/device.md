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

## Import

Device can be imported using an id made up of `<device_name>`, e.g.

```
$ terraform import wallix-bastion_device.server1 server1
```
