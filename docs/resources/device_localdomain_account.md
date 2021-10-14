# wallix-bastion_device_localdomain_account Resource

Provides a account linked to device_localdomain resource.

## Example Usage

```hcl
# Configure a account into local domain of a device
resource wallix-bastion_device_localdomain_account srv1adm {
  device_id     = "xxxxxxxx"
  domain_id     = "yyyyyyy"
  account_name  = "admin"
  account_login = "admin"
}
```

## Argument Reference

The following arguments are supported:

- **device_id** (Required, String, Forces new resource)  
  ID of device.
- **domain_id** (Required, String, Forces new resource)  
  ID of localdomain.
- **account_name** (Required, String)  
  The account name.
- **account_login** (Required, String)  
  The account login.
- **auto_change_password** (Optional, Boolean)  
  Automatically change the password.
- **auto_change_ssh_key** (Optional, Boolean)  
  Automatically change the ssh key.
- **certificate_validity** (Optional, String)  
  The validity duration of the signed ssh public key in the case a Certificate Authority is defined
  for the account's domain.
- **checkout_policy** (Optional, String)  
  The account checkout policy.  
  Default to `default`.
- **description** (Optional, String)  
  The account description.
- **services** (Optional, List of String)  
  The account services.

## Attribute Reference

- **id** (String)  
  Internal id of localdomain account in bastion.
- **credentials** (List of Block)  
  The account credentials.
  - **id** (String)  
    Internal id of credential.
  - **public_key** (String)  
    The account public key (if `type` = `ssh_key`).
  - **type** (String)  
    The credential type.
- **domain_password_change** (Boolean)  
  True if the password change is configured on the domain.

## Import

Account linked to device_localdomain can be imported using an id made up
of `<device_id>/<domain_id>/<account_name>`, e.g.

```shell
terraform import wallix-bastion_device_localdomain_account.srv1adm xxxxxxxx/yyyyyyy/admin
```
