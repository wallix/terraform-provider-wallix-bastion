# wallix-bastion_user Resource

Provides a user resource.

## Argument Reference

The following arguments are supported:

* `user_name` - (Required, Forces new resource)(`String`) The user name.
* `email` - (Required)(`String`) The email address.
* `profile` - (Required)(`String`) The user profile.
* `user_auths` - (Required)(`ListOfString`) The authentication procedures(s).
* `certificate_dn` - (Optional)(`String`) The certificate DN (for X509 authentication).
* `display_name` - (Optional)(`String`) The displayed name.
* `expiration_date` - (Optional)(`String`) Account expiration date/time (format: "yyyy-mm-dd hh:mm").
* `force_change_pwd` - (Optional)(`Bool`) Force password change. **Only used when create resource**
* `groups` - (Optional)(`ListOfString`) The groups containing this user. **It's a attributes when not set**
* `ip_source` - (Optional)(`String`) The source IP to limit access. Format is a comma-separated list of IPv4 addresses, subnets or ranges.
* `is_disabled` - (Optional)(`Bool`) Account is disabled.
* `password` - (Optional)(`String`) The password. **Only used when create resource**
* `preferred_language` - (Optional)(`String`) The preferred language. **Only used when create resource** Need to be 'de', 'en', 'es', 'fr' or 'ru'.
* `ssh_public_key` - (Optional)(`String`) The SSH public key.

## Attribute Reference

* `id` - (`String`) = `user_name`

## Import

User can be imported using an id made up of `<user_name>`, e.g.

```
$ terraform import wallix-bastion_user.toto toto
```
