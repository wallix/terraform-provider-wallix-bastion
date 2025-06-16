# wallix-bastion_authdomain_saml Resource

Provides a Other IDPs/SAML auth domain resource.

## Example Usage

```hcl
# Configure a Other IDPs/SAML auth domain
resource "wallix-bastion_authdomain_saml" "example_com" {
  domain_name          = "example.com"
  auth_domain_name     = "example.com"
  external_auths       = "saml"
  default_language     = "fr"
  default_email_domain = "example.com"
  label                = "SAML"
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
- **label** (Required, String)  
  The label to display on the login page.
- **description** (Optional, String)  
  The domain description.
- **force_authn** (Optional, Boolean)  
  Force authentication on IdP at each login.
- **is_default** (Optional, Boolean)  
  The domain is used by default.
- **secondary_auth** (Optional, List of String)  
  The secondary authentications methods for the auth domain

## Attribute Reference

- **id** (String)  
  Internal id of auth domain in bastion.
- **idp_initiated_url** (String)  
  URL used in Identity Provider (IdP) initiated Single Sign-On (SSO) flows.

## Import

Other IDPs/SAML auth domain can be imported using an id made up of `<domain_name>`, e.g.

```shell
terraform import wallix-bastion_authdomain_saml.example_com example.com
```
