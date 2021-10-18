# wallix-bastion_device_localdomain Resource

Provides a localdomain resource linked to device.

## Example Usage

```hcl
# Configure a local domain on device
resource wallix-bastion_device_localdomain srv1dom {
  device_id   = "xxxxxxxx"
  domain_name = "domlocal"
}
```

## Argument Reference

The following arguments are supported:

- **device_id** (Required, String, Forces new resource)  
  ID of device.
- **domain_name** (Required, String)  
  The domain name.
- **admin_account** (Optional, String, **Not used when create**)  
  The administrator account used to change passwords on this domain.  
  Need `enable_password_change` to true.
- **ca_private_key** (Optional, String, **Value can't refresh**)  
  CA certificate.  
  The ssh private key of the signing authority for the ssh keys for accounts in the domain.  
  Special values are allowed to automatically generate SSH key:
  `generate:RSA_1024`, `generate:RSA_2048`, `generate:RSA_4096`, `generate:RSA_8192`,
  `generate:DSA_1024`, `generate:ECDSA_256`, `generate:ECDSA_384`, `generate:ECDSA_521`, `generate:ED25519`.
- **description** (Optional, String)  
  The domain description.
- **enable_password_change** (Optional, Boolean)  
  Enable the change of password on this domain.  
  `password_change_policy`, `password_change_plugin` and `password_change_plugin_parameters` need to
  be set.
- **passphrase** (Optional, String, **Value can't refresh**)  
  The passphrase that was used to encrypt the private key.  
  If provided, it must be between 4 and 1024 characters long.
- **password_change_policy** (Optional, String)  
  The name of password change policy for this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin** (Optional, String)  
  The name of plugin used to change passwords on this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin_parameters** (Optional, String, Sensitive, **Value can't refresh**)  
  Parameters for the plugin used to change credentials.  
  Need to be a valid JSON.  
  Need `enable_password_change` to true.

## Attribute Reference

- **id** (String)  
  Internal id of local domain in bastion.
- **ca_public_key** (String)  
  The ssh public key of the signing authority for the ssh keys for accounts in the domain.

## Import

Localdomain linked to device can be imported using an id made up of `<device_id>/<domain_name>`, e.g.

```shell
terraform import wallix-bastion_device_localdomain.srv1dom xxxxxxxx/domlocal
```
