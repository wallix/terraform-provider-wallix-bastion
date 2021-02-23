# wallix-bastion_device_service Resource

Provides a service resource linked to device.

## Argument Reference

The following arguments are supported:

* `device_id ` - (Required, Forces new resource)(`String`) ID of device.
* `service_name` - (Required, Forces new resource)(`String`) The service name.
* `connection_policy` - (Required)(`String`) The connection policy name.
* `port` - (Required)(`Int`) The port number.
* `protocol` - (Required, Forces new resource)(`String`) The protocol. Need to be 'SSH', 'RAWTCPIP', 'RDP', 'RLOGIN', 'TELNET' or 'VNC'.
* `global_domains` - (Optional)(`ListOfString`) The global domains names.
* `subprotocols` - (Optional)(`ListOfString`) The sub protocols for 'SSH', 'RDP' protocol.

## Attribute Reference

* `id` - (`String`) Internal id of service in bastion.

## Import

Service linked to device can be imported using an id made up of `<device_id>/<service_name>`, e.g.

```
$ terraform import wallix-bastion_device_service.srv1svc xxxxxxxx/svc
```
