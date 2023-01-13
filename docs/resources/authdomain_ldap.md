# wallix-bastion_authdomain_ldap Resource

Provides a LDAP auth domain resource.

## Example Usage

```hcl
# Configure a LDAP auth domain
resource "wallix-bastion_authdomain_ldap" "example_com" {
  domain_name          = "example.com"
  auth_domain_name     = "example.com"
  external_auths       = "server1"
  default_language     = "fr"
  default_email_domain = "example.com"
}
```

## Argument Reference

The following arguments are supported:

- **domain_name** (Required, String, Forces new resource)  
  The domain name.
- **auth_domain_name** (Required, String)  
  The auth domain name.
- **default_email_domain** (Required, String)  
  The default email domain.
- **default_language** (Required, String)  
  The default language.  
  Need to be `de`, `en`, `es`, `fr` or `ru`.
- **external_auths** (Required, List of String)  
  The external authentications.
- **description** (Optional, String)  
  The domain description.
- **check_x509_san_email** (Optional, Boolean)  
  Match the X509v3 SAN email.
- **display_name_attribute** (Optional, String)  
  The display name attribute.
- **email_attribute** (Optional, String)  
  The email attribute.
- **group_attribute** (Optional, String)  
  The group attribute.
- **is_default** (Optional, Boolean)  
  The domain is used by default.
- **language_attribute** (Optional, String)  
  The language attribute.
- **pubkey_attribute** (Optional, String)  
  The SSH public key attribute.
- **san_domain_name** (Optional, String)  
  The domain name to match SAN email (only for AD server).
- **secondary_auth** (Optional, List of String)  
  The secondary authentications methods for the auth domain
- **x509_condition** (Optional, String)  
  Condition to match a LDAP domain with the X509 certificate variables (only for LDAP server).
- **x509_search_filter** (Optional, String)  
  LDAP search filter for X509 authentication (only for LDAP server).

## Attribute Reference

- **id** (String)  
  Internal id of auth domain in bastion.

## Import

LDAP auth domain can be imported using an id made up of `<domain_name>`, e.g.

```shell
terraform import wallix-bastion_authdomain_ldap.example_com example.com
```
