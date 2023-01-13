# wallix-bastion_connection_policy Resource

Provides a connection_policy resource.

## Example Usage

```hcl
# Configure a connection policy
resource "wallix-bastion_connection_policy" "pol" {
  connection_policy_name = "example"
  protocol               = "RAWTCPIP"
  options = jsonencode({
    general = {}
  })
}
```

## Argument Reference

The following arguments are supported:

- **connection_policy_name** (Required, String)  
  The connection policy name.
- **protocol** (Required, String)  
  The connection policy protocol.
  Need to be `SSH`, `RAWTCPIP`, `RDP`, `RLOGIN`, `TELNET` or `VNC`.
- **description** (Optional, String)  
  The connection policy description.
- **authentication_methods** (Optional, Set of String)  
  The allowed authentication methods.
- **options** (Optional, String)  
  Options for the connection policy.  
  Need to be a valid JSON.

## Attribute Reference

- **id** (String)  
  Internal id of connection policy in bastion.

## Import

Connection policy can be imported using an id made up of `<connection_policy_name>`, e.g.

```shell
terraform import wallix-bastion_connection_policy.pol example
```
