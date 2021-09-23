# wallix-bastion_domain_account_credential Resource

Provides a credential linked to domain_account resource.

## Argument Reference

The following arguments are supported:

- **domain_id** (Required, String, Forces new resource)  
  ID of domain.
- **account_id** (Required, String, Forces new resource)  
  ID of account.
- **type** (Required, String, Forces new resource)  
  The credential type.  
  Need to be `password` or `ssh_key`.
- **passphrase** (Optional, String, Sensitive, **Value can't refresh**)  
  The passphrase for the private key (only for an encrypted private key).
- **password** (Optional, String, Sensitive, **Value can't refresh**)  
  The account password.
- **private_key** (Optional, String, Sensitive, **Value can't refresh**)  
  The account private key.  
  Special values are allowed to automatically generate SSH key: `generate:RSA_1024`, `generate:RSA_2048`, `generate:RSA_4096`, `generate:RSA_8192`, `generate:DSA_1024`, `generate:ECDSA_256`, `generate:ECDSA_384`, `generate:ECDSA_521`, `generate:ED25519`.

## Attribute Reference

- **id** (String)  
  Internal id of domain account credential in bastion.
- **public_key** (String)  
  The account public key.

## Import

Credential linked to domain_account can be imported using an id made up of `<domain_id>/<account_id>/<type>`, e.g.

```shell
terraform import wallix-bastion_domain_account_credential.dom1admpass xxxxxxxx/yyyyyyy/password
```
