# wallix-bastion_domain_account Resource

Provides a account linked to domain resource.

## Argument Reference

The following arguments are supported:

* `domain_id` - (Required, Forces new resource)(`String`) ID of domain.
* `account_name` - (Required)(`String`) The account name.
* `account_login` - (Required)(`String`) The account login.
* `auto_change_password` - (Optional)(`Bool`) Automatically change the password.
* `auto_change_ssh_key` - (Optional)(`Bool`) Automatically change the ssh key.
* `certificate_validity` - (Optional)(`String`) The validity duration of the signed ssh public key in the case a Certificate Authority is defined for the account's domain.
* `checkout_policy` - (Optional)(`String`) The account checkout policy.
* `description` - (Optional)(`String`) The account description.
* `resources` - (Optional)(`ListOfString`) The account resources. Format is device:service or application:APP.

## Attribute Reference

* `id` - (`String`) Internal id of domain account in bastion.
* `credentials` - (`ListOfBlock`) The account credentials.
  * `id` - (`String`) Internal id of credential.
  * `public_key` - (`String`) The account public key.
  * `type` - (`String`) The credential type.
* `domain_password_change` - (`Bool`) True if the password change is configured on the domain.


## Import

Account linked to domain can be imported using an id made up of `<domain_id>/<account_name>`, e.g.

```
$ terraform import wallix-bastion_domain_account.dom1adm xxxxxxxx/admin
```
