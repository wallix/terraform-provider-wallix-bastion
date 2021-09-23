# wallix-bastion_user Resource

Provides a user resource.

## Argument Reference

The following arguments are supported:

- **user_name** (Required, String, Forces new resource)  
  The user name.
- **email** (Required, String)  
  The email address.
- **profile** (Required, String)  
  The user profile.
- **user_auths** (Required, List of String)  
  The authentication procedures(s).
- **certificate_dn** (Optional, String)  
  The certificate DN (for X509 authentication).
- **display_name** (Optional, String)  
  The displayed name.
- **expiration_date** (Optional, String)  
  Account expiration date/time.  
  Format: `yyyy-mm-dd hh:mm`.
- **force_change_pwd** (Optional, Boolean, **Only used when create resource**)  
  Force password change.
- **groups** (Optional, List of String, **It's a attributes when not set**)  
  The groups containing this user.
- **ip_source** (Optional, String)  
  The source IP to limit access.  
  Format is a comma-separated list of IPv4 addresses, subnets or ranges.
- **is_disabled** (Optional, Boolean)  
  Account is disabled.
- **password** (Optional, String, Sensitive, **Only used when create resource**)  
  The password.
- **preferred_language** (Optional, String, **Only used when create resource**)  
  The preferred language.  
  Need to be `de`, `en`, `es`, `fr` or `ru`.
- **ssh_public_key** (Optional, String)  
  The SSH public key.

## Attribute Reference

- **id** (String)  
  ID of resource = `user_name`

## Import

User can be imported using an id made up of `<user_name>`, e.g.

```shell
terraform import wallix-bastion_user.toto toto
```
