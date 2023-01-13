# wallix-bastion_authdomain_azuread Resource

Provides a AzureAD auth domain resource.

## Example Usage

```hcl
# Configure a AzureAD auth domain
resource "wallix-bastion_authdomain_azuread" "example_com" {
  domain_name          = "example.com"
  auth_domain_name     = "example.com"
  external_auths       = "saml"
  default_language     = "fr"
  default_email_domain = "example.com"
  client_id            = "xxx"
  entity_id            = "yyy"
  label                = "AzureAD"
}
```

## Argument Reference

The following arguments are supported:

- **domain_name** (Required, String, Forces new resource)  
  The domain name.
- **auth_domain_name** (Required, String)  
  The auth domain name.
- **client_id** (Required, String)  
  The application (client) ID.
- **default_email_domain** (Required, String)  
  The default email domain.
- **default_language** (Required, String)  
  The default language.  
  Need to be `de`, `en`, `es`, `fr` or `ru`.
- **entity_id** (Required, String)  
  The entity (tenant) ID.
- **external_auths** (Required, List of String)  
  The external authentications.
- **label** (Required, String)  
  The label to display on the login page.
- **certificate** (Optional, String, Sensitive, **Value can't refresh**)  
  The client certificate.
- **client_secret** (Optional, String, Sensitive, **Value can't refresh**)  
  The client secret.
- **description** (Optional, String)  
  The domain description.
- **is_default** (Optional, Boolean)  
  The domain is used by default.
- **passphrase** (Optional, String, Sensitive, **Value can't refresh**)  
  The passphrase (if the private key is encrypted).
- **private_key** (Optional, String, Sensitive, **Value can't refresh**)  
  The client private key.
- **secondary_auth** (Optional, List of String)  
  The secondary authentications methods for the auth domain

## Attribute Reference

- **id** (String)  
  Internal id of auth domain in bastion.

## Import

AzureAD auth domain can be imported using an id made up of `<domain_name>`, e.g.

```shell
terraform import wallix-bastion_authdomain_azuread.example_com example.com
```
