# wallix-bastion_application Resource

Provides a application resource.

## Argument Reference

The following arguments are supported:

* `application_name` - (Required)(`String`) The application name.
* `connection_policy` - (Required)(`String`) The connection policy name.
* `paths` - (Required)(`NestedBlock`) Need to be specified multiple times for each target in cluster or once if target is an application.
* `target` - (Required)(`String`) The application target/cluster name.
* `description` - (Optional)(`String`) The application description.
* `global_domains` - (Optional)(`ListOfString`) The global domains names.
* `parameters` - (Optional)(`String`) The application parameters.

## Attribute Reference

* `id` - (`String`) Internal id of application in bastion.
* `local_domains` - (`ListOfBlock`) List of localdomain.
  * `id` - (`String`) Internal id of local domain in bastion.
  * `domain_name` - (`String`) The domain name.
  * `description` - (`String`) The domain description.
  * `enable_password_change` - (`Bool`) Enable the change of password on this domain.
  * `password_change_policy` - (`String`) The name of password change policy for this domain.
  * `password_change_plugin` - (`String`) The name of plugin used to change passwords on this domain.
  * `password_change_plugin_parameters` - (`String`) Parameters for the plugin used to change credentials.

## Import

Application can be imported using an id made up of `<application_name>`, e.g.

```
$ terraform import wallix-bastion_application.app1 app1
```
