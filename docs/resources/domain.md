# wallix-bastion_domain Resource

Provides a global domain resource.

## Example Usage

```hcl
# Configure a global domain
resource wallix-bastion_domain cmpdom {
  domain_name = "globdom"
}
```

## Argument Reference

The following arguments are supported:

- **domain_name** (Required, String)  
  The domain name.
- **domain_real_name** (Optional, String)  
  The domain name used for connection to a target.
- **admin_account** (Optional, String, **Not used when create**)  
  The administrator account used to change passwords on this domain.  
  Need `enable_password_change` to true.
- **ca_private_key** (Optional, String, Sensitive, **Value can't refresh**)  
  CA certificate.  
  The ssh private key of the signing authority for the ssh keys for accounts in the domain.  
  Special values are allowed to automatically generate SSH key: `generate:RSA_1024`, `generate:RSA_2048`, `generate:RSA_4096`, `generate:RSA_8192`, `generate:DSA_1024`, `generate:ECDSA_256`, `generate:ECDSA_384`, `generate:ECDSA_521`, `generate:ED25519`.  
  Conflict with `vault_plugin`.
- **description** (Optional, String)  
  The domain description.
- **enable_password_change** (Optional, Boolean)  
  Enable the change of password on this domain.  
  `password_change_policy`, `password_change_plugin` and `password_change_plugin_parameters` need to be set.  
  Conflict with `vault_plugin`.
- **passphrase** (Optional, String, Sensitive, **Value can't refresh**)  
  The passphrase that was used to encrypt the private key. If provided, it must be between 4 and 1024 characters long.
- **password_change_policy** (Optional, String)  
  The name of password change policy for this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin** (Optional, String, Sensitive, **Value can't refresh**)  
  The name of plugin used to change passwords on this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin_parameters** (Optional, String)  
  Parameters for the plugin used to change credentials.  
  Need to be a valid JSON.  
  Need `enable_password_change` to true.
- **vault_plugin** (Optional, String, Force new resource)  
  The name of vault plugin used to manage all accounts defined on this domain.  
  Conflict with `enable_password_change` and `ca_private_key`.
  Need `vault_plugin_parameters` to be set.
- **vault_plugin_parameters** (Optional, String, Sensitive, **Value can't refresh**)  
  Parameters for the vault plugin.  
  Need to be a valid JSON.
  Need `vault_plugin` to be set.  

## Attribute Reference

- **id** (String)  
  Internal id of domain in bastion.
- **ca_public_key** (String)  
  The ssh public key of the signing authority for the ssh keys for accounts in the domain.

## Import

Domain can be imported using an id made up of `<domain_name>`, e.g.

```shell
terraform import wallix-bastion_domain.cmpdom globdom
```
