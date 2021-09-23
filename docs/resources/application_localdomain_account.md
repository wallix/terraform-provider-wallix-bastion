# wallix-bastion_application_localdomain_account Resource

Provides a account linked to application_localdomain resource.

## Argument Reference

The following arguments are supported:

- **application_id** (Required, String, Forces new resource)  
  ID of application.
- **domain_id** (Required, String, Forces new resource)  
  ID of localdomain.
- **account_name** (Required, String)  
  The account name.
- **account_login** (Required, String)  
  The account login.
- **auto_change_password** (Optional, Boolean)  
  Automatically change the password.
- **checkout_policy** (Optional, String)  
  The account checkout policy.  
  Default to `default`.
- **description** (Optional, String)  
  The account description.
- **password** (Optional, String, Sensitive, **Value can't refresh**)  
  The account password.

## Attribute Reference

- **id** (String)  
  Internal id of localdomain account in bastion.
- **domain_password_change** (Boolean)  
  True if the password change is configured on the domain.

## Import

Account linked to application_localdomain can be imported using an id made up of `<application_id>/<domain_id>/<account_name>`, e.g.

```shell
terraform import wallix-bastion_application_localdomain_account.app1adm xxxxxxxx/yyyyyyy/admin
```
