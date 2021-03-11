# wallix-bastion_application_localdomain Resource

Provides a localdomain resource linked to application.

## Argument Reference

The following arguments are supported:

* `application_id` - (Required, Forces new resource)(`String`) ID of application.
* `domain_name` - (Required)(`String`) The domain name.
* `admin_account` - (Optional)(`String`) The administrator account used to change passwords on this domain. Need `enable_password_change` to true. **Not used when create**
* `description` - (Optional)(`String`) The domain description.
* `enable_password_change` - (Optional)(`Bool`) Enable the change of password on this domain. RequiredWith arguments : `password_change_policy` and `password_change_plugin`.
* `password_change_policy` - (Optional,Required)(`String`) The name of password change policy for this domain.  Need `enable_password_change` to true.
* `password_change_plugin` - (Optional)(`String`) The name of plugin used to change passwords on this domain.  Need `enable_password_change` to true.
* `password_change_plugin_parameters` - (Optional)(`String`) Parameters for the plugin used to change credentials. Need to be a valid JSON. Need `enable_password_change` to true. **Value can't refresh**  

## Attribute Reference

* `id` - (`String`) Internal id of local domain in bastion.

## Import

Localdomain linked to application can be imported using an id made up of `<application_id>/<domain_name>`, e.g.

```
$ terraform import wallix-bastion_application_localdomain.app1dom xxxxxxxx/domlocal
```
