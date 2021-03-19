# wallix-bastion_application_localdomain_account Resource

Provides a account linked to application_localdomain resource.

## Argument Reference

The following arguments are supported:

* `application_id` - (Required, Forces new resource)(`String`) ID of application.
* `domain_id` - (Required, Forces new resource)(`String`) ID of localdomain.
* `account_name` - (Required)(`String`) The account name.
* `account_login` - (Required)(`String`) The account login.
* `auto_change_password` - (Optional)(`Bool`) Automatically change the password.
* `checkout_policy` - (Optional)(`String`) The account checkout policy.
* `description` - (Optional)(`String`) The account description.
* `password` - (Optional)(`String`) The account password.

## Attribute Reference

* `id` - (`String`) Internal id of localdomain account in bastion.
* `domain_password_change` - (`Bool`) True if the password change is configured on the domain.


## Import

Account linked to application_localdomain can be imported using an id made up of `<application_id>/<domain_id>/<account_name>`, e.g.

```
$ terraform import wallix-bastion_application_localdomain_account.app1adm xxxxxxxx/yyyyyyy/admin
```
