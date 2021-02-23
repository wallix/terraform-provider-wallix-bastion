# wallix-bastion_device_localdomain Resource

Provides a localdomain resource linked to device.

## Argument Reference

The following arguments are supported:

* `device_id` - (Required, Forces new resource)(`String`) ID of device.
* `domain_name` - (Required)(`String`) The domain name.
* `admin_account` - (Optional)(`String`) The administrator account used to change passwords on this domain. Need `enable_password_change` to true.
* `ca_private_key` - (Optional)(`String`) CA certificate. The ssh private key of the signing authority for the ssh keys for accounts in the domain. Special values are allowed to automatically generate SSH key: "generate:RSA_1024", "generate:RSA_2048", "generate:RSA_4096", "generate:RSA_8192", "generate:DSA_1024", "generate:ECDSA_256", "generate:ECDSA_384", "generate:ECDSA_521", "generate:ED25519". **Value can't refresh**
* `description` - (Optional)(`String`) The domain description.
* `enable_password_change` - (Optional)(`Bool`) Enable the change of password on this domain. RequiredWith arguments : `password_change_policy` and `password_change_plugin`.
* `passphrase` - (Optional)(`String`) The passphrase that was used to encrypt the private key. If provided, it must be between 4 and 1024 characters long. **Value can't refresh**
* `password_change_policy` - (Optional,Required)(`String`) The name of password change policy for this domain.  Need `enable_password_change` to true.
* `password_change_plugin` - (Optional)(`String`) The name of plugin used to change passwords on this domain.  Need `enable_password_change` to true.
* `password_change_plugin_parameters` - (Optional)(`String`) Parameters for the plugin used to change credentials. Need to be a valid JSON. Need `enable_password_change` to true. **Value can't refresh**  

## Attribute Reference

* `id` - (`String`) Internal id of local domain in bastion.
* `ca_public_key` - (`String`) The ssh public key of the signing authority for the ssh keys for accounts in the domain.

## Import

Localdomain linked to device can be imported using an id made up of `<device_id>/<domain_name>`, e.g.

```
$ terraform import wallix-bastion_device_localdomain.srv1dom xxxxxxxx/domlocal
```
