# wallix-bastion_externalauth_radius Resource

Provides a Radius externaulauth resource.

## Example Usage

```hcl
# Configure a radius external authentication
resource wallix-bastion_externalauth_radius server1 {
  authentication_name = "server1"
  host                = "server1"
  port                = 1813
  secret              = "aSecret"
  timeout             = 10
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
- **secret** (Required, String, Sensitive)  
  The secret.
- **timeout** (Required, Number)  
  Radius timeout.
- **description** (Optional, String)  
  Description of the authentication.
- **use_primary_auth_domain** (Optional, Boolean)  
  Use the primary auth domain.

## Attribute Reference

- **id** (String)  
  Internal id of externalauth in bastion.

## Import

Radius externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```shell
terraform import wallix-bastion_externalauth_radius.server1 server1
```
