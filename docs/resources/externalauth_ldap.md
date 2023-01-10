# wallix-bastion_externalauth_ldap Resource

Provides a LDAP externaulauth resource.

## Example Usage

```hcl
# Configure a LDAP external authentication
resource "wallix-bastion_externalauth_ldap" "server1" {
  authentication_name = "server1"
  cn_attribute        = "sAMAccountName"
  host                = "server1"
  ldap_base           = "OU=FR,DC=test,DC=com"
  login_attribute     = "sAMAccountName"
  port                = 636
  timeout             = 10
  is_ssl              = true
  is_anonymous_access = true
}
```

## Argument Reference

The following arguments are supported:

- **authentication_name** (Required, String)  
  The authentication name.
- **cn_attribute** (Required, String)  
  The username attribute.
- **host** (Required, String)  
  The host name.
- **ldap_base** (Required, String)  
  The LDAP base scheme.
- **login_attribute** (Required, String)  
  The login attribute.
- **port** (Required, Number)  
  The port number.
- **timeout** (Required, Number)  
  LDAP timeout.
- **ca_certificate** (Optional, String)  
  CA certificate.
- **certificate** (Optional, String)  
  Client certificate.
- **description** (Optional, String)  
  Description of the authentication.
- **is_active_directory** (Optional, Boolean)  
  This LDAP uses an active directory.
- **is_anonymous_access** (Optional, Boolean)  
  The user is anonymous.
- **is_protected_user** (Optional, Boolean)  
  The AD user is protected.
- **is_ssl** (Optional, Boolean)  
  This LDAP is secure (with SSL/TLS).
- **is_starttls** (Optional, Boolean)  
  This LDAP uses STARTTLS.
- **login** (Optional, String)  
  The login.  
  Required if `is_anonymous_access` = `false`.
- **password** (Optional, String, Sensitive, **Value can't refresh**)  
  The password.  
  Required if `is_anonymous_access` = `false`.
- **private_key** (Optional, String)  
  Client key.
- **use_primary_auth_domain** (Optional, Boolean)  
  Use the primary auth domain.

## Attribute Reference

- **id** (String)  
  Internal id of externalauth in bastion.

## Import

LDAP externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```shell
terraform import wallix-bastion_externalauth_ldap.server1 server1
```
