# wallix-bastion_domain Data Source

Get information on a global domain resource.

## Example Usage

```hcl
data "wallix-bastion_domain" "globDomain" {
  domain_name = "globDomain"
}
```

## Argument Reference

The following arguments are supported:

- **domain_name** (Required, String)  
  The domain name.

## Attribute Reference

- **id** (String)  
  Internal id of domain in bastion.
- **domain_real_name** (String)  
  The domain name used for connection to a target.
- **admin_account** (String)  
  The administrator account used to change passwords on this domain.
- **ca_public_key** (String)  
  The ssh public key of the signing authority for the ssh keys for accounts in the domain.
- **description** (String)  
  The domain description.
- **enable_password_change** (Boolean)  
  Enable the change of password on this domain.
- **password_change_policy** (String)  
  The name of password change policy for this domain.
- **password_change_plugin** (String)  
  The name of plugin used to change passwords on this domain.
- **vault_plugin** (String)  
  The name of vault plugin used to manage all accounts defined on this domain.
