# wallix-bastion_device_localdomain_account_credential Resource

Provides a credential linked to device_localdomain_account resource.

## Argument Reference

The following arguments are supported:

* `device_id` - (Required, Forces new resource)(`String`) ID of device.
* `domain_id` - (Required, Forces new resource)(`String`) ID of localdomain.
* `account_id` - (Required, Forces new resource)(`String`) ID of account.
* `type` - (Required, Forces new resource)(`String`) The credential type. Need to be 'password' or 'ssh_key'.
* `passphrase` - (Optional)(`String`) The passphrase for the private key (only for an encrypted private key). **Value can't refresh**
* `password` - (Optional)(`String`) The account password. **Value can't refresh**
* `private_key` - (Optional)(`String`) The account private key. Special values are allowed to automatically generate SSH key: "generate:RSA_1024", "generate:RSA_2048", "generate:RSA_4096", "generate:RSA_8192", "generate:DSA_1024", "generate:ECDSA_256", "generate:ECDSA_384", "generate:ECDSA_521", "generate:ED25519". **Value can't refresh**

## Attribute Reference

* `id` - (`String`) Internal id of local domain in bastion.
* `public_key` - (`String`) The account public key.

## Import

Credential linked to device_localdomain_account can be imported using an id made up of `<device_id>/<domain_id>/<account_id>/<type>`, e.g.

```
$ terraform import wallix-bastion_device_localdomain_account_credential.srv1admpass xxxxxxxx/yyyyyyy/zzzzz/password
```
