# wallix-bastion_externalauth_tacacs Resource

Provides a Tacacs+ externaulauth resource.

## Example Usage

```hcl
# Configure a tacacs external authentication
resource "wallix-bastion_externalauth_tacacs" "server1" {
  authentication_name = "server1"
  host                = "server1"
  port                = 49
  secret              = "aSecret"
}
```

## Argument Reference

The following arguments are supported:

- **authentication_name** (Required, String)  
  The authentication name.
- **host** (Required, String)  
  The host name.
- **port** (Required, Number)  
  The port number.
- **secret** (Required, String, Sensitive, **Value can't refresh**)  
  The secret.
- **description** (Optional, String)  
  Description of the authentication.
- **use_primary_auth_domain** (Optional, Boolean)  
  Use the primary auth domain.

## Attribute Reference

- **id** (String)  
  Internal id of externalauth in bastion.

## Import

Tacacs+ externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```shell
terraform import wallix-bastion_externalauth_tacacs.server1 server1
```
