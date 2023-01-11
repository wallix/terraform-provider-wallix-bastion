# wallix-bastion_externalauth_kerberos Resource

Provides a Kerberos externaulauth resource.

## Example Usage

```hcl
# Configure a tacacs external authentication
resource "wallix-bastion_externalauth_kerberos" "server1" {
  authentication_name = "server1"
  host                = "server1"
  ker_dom_controller  = "controller"
  port                = 88
  keytab              = filebase64("keytab")
}
```

## Argument Reference

The following arguments are supported:

- **authentication_name** (Required, String)  
  The authentication name.
- **host** (Required, String)  
  The host name.
- **ker_dom_controller** (Required, String)  
  Kerberos domain controller whose role is torecognizes the tickets issued bythe Key Distribution Center.
- **port** (Required, Number)  
  The port number.
- **kerberos_password** (Optional, Boolean, Force new resource)  
  Use KERBEROS-PASSWORD protocol.
- **description** (Optional, String)  
  Description of the authentication.
- **keytab** (Optional, String)  
  The keytab file, containing pairs of principal and encrypted keys.  
  The content of the file needed must be converted to base64 before being sent.
- **login_attribute** (Optional, String)  
  The login attribute.
- **use_primary_auth_domain** (Optional, Boolean)  
  Use the primary auth domain.

## Attribute Reference

- **id** (String)  
  Internal id of externalauth in bastion.

## Import

Kerberos externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```shell
terraform import wallix-bastion_externalauth_kerberos.server1 server1
```
