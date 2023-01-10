# wallix-bastion_device_service Resource

Provides a service resource linked to device.

## Example Usage

```hcl
# Configure a service on device
resource "wallix-bastion_device_service" "srv1svc" {
  device_id         = "xxxxxxxx"
  service_name      = "svc"
  connection_policy = "SSH"
  port              = 22
  protocol          = "SSH"
  subprotocols      = ["SSH_SHELL_SESSION"]
}
```

## Argument Reference

The following arguments are supported:

- **device_id** (Required, String, Forces new resource)  
  ID of device.
- **service_name** (Required, String, Forces new resource)  
  The service name.
- **connection_policy** (Required, String)  
  The connection policy name.
- **port** (Required, Number)  
  The port number.
- **protocol** (Required, String, Forces new resource)  
  The protocol.  
  Need to be `SSH`, `RAWTCPIP`, `RDP`, `RLOGIN`, `TELNET` or `VNC`.
- **global_domains** (Optional, List of String, **It's an attribute when not set**)  
  The global domains names.
- **subprotocols** (Optional, List of String)  
  The sub protocols for `SSH`, `RDP` protocol.

## Attribute Reference

- **id** (String)  
  Internal id of service in bastion.

## Import

Service linked to device can be imported using an id made up of `<device_id>/<service_name>`, e.g.

```shell
terraform import wallix-bastion_device_service.srv1svc xxxxxxxx/svc
```
