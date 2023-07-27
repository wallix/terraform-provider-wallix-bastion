# wallix-bastion_connection_message Resource

Update a connection message

-> **Note:** `Create` operation overrides the current message in the bastion.
`Delete` operation has no effect.

## Example Usage

```hcl
# Update a connection message
resource "wallix-bastion_connection_message" "SecondaryEn" {
  message_name = "motd_en"
  message      = <<EOT
You are hereby informed and acknowledge that your actions may be recorded, retained and audited in accordance with your organization security policy.
Please contact your WALLIX Bastion administrator for further information.
EOT
}
```

## Argument Reference

The following arguments are supported:

- **message_name** (Required, String, Forces new resource)  
  The connection message name.  
  Need to be `login_en`, `login_fr`, `login_de`, `login_es`, `login_ru`,
  `motd_en`, `motd_fr`, `motd_de`, `motd_es` or `motd_ru`.
- **message** (Required, String)  
  Content of the message.

## Attribute Reference

- **id** (String)  
  ID of resource = `message_name`

## Import

Connection message can be imported using an id made up of `<message_name>`, e.g.

```shell
terraform import wallix-bastion_connection_message.SecondaryEn motd_en
```
