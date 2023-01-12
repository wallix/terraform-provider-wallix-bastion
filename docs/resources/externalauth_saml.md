# wallix-bastion_externalauth_saml Resource

Provides a SAML externaulauth resource.

## Example Usage

```hcl
# Configure a SAML external authentication
resource "wallix-bastion_externalauth_saml" "example_com" {
  authentication_name = "example_com"
  idp_metadata        = file("idp_metadata.xml")
  timeout             = 30
}
```

## Argument Reference

The following arguments are supported:

- **authentication_name** (Required, String)  
  The authentication name.
- **idp_metadata** (Required, String)  
  Identity Provider metadata (XML format).
- **timeout** (Required, Number)  
  SAML request timeout.
- **certificate** (Optional, String, Sensitive, **Value can't refresh**)  
  The certificate of the Service Provider.
- **description** (Optional, String)  
  Description of the authentication.
- **passphrase** (Optional, String, Sensitive, **Value can't refresh**)  
  The Passphrase for the private key (only for an encrypted private key).
- **private_key** (Optional, String, Sensitive, **Value can't refresh**)  
  The private key of the Service Provider.

## Attribute Reference

- **id** (String)  
  Internal id of externalauth in bastion.
- **idp_entity_id** (String)  
  Identifier of the IdP entity.
- **saml_request_url** (String)  
  Single Sign-On URL.
- **saml_request_method** (String)  
  Single Sign-On binding.
- **sp_metadata** (String)  
  Service Provider metadata (XML format).
- **sp_entity_id** (String)  
  Identifier of the SP entity.
- **sp_assertion_consumer_service** (String)  
  Assertion Consumer Service URL (Service Provider).
- **sp_single_logout_service** (String)  
  Single Logout Service URL (Service Provider).

## Import

SAML externalauth can be imported using an id made up of `<authentication_name>`, e.g.

```shell
terraform import wallix-bastion_externalauth_saml.example_com example_com
```
