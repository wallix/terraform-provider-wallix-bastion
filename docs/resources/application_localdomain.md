# wallix-bastion_application_localdomain Resource

Provides a localdomain resource linked to application.

## Argument Reference

The following arguments are supported:

- **application_id** (Required, String, Forces new resource)  
  ID of application.
- **domain_name** (Required, String)  
  The domain name.
- **admin_account** (Optional, String,  **Not used when create**)  
  The administrator account used to change passwords on this domain.  
  Need `enable_password_change` to true.  
- **description** (Optional, String)  
  The domain description.
- **enable_password_change** (Optional, Boolean)  
  Enable the change of password on this domain.  
  `password_change_policy`, `password_change_plugin` and `password_change_plugin_parameters` need to be set.
- **password_change_policy** (Optional, String)  
  The name of password change policy for this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin** (Optional, String)  
  The name of plugin used to change passwords on this domain.  
  Need `enable_password_change` to true.
- **password_change_plugin_parameters** (Optional, String, Sensitive, **Value can't refresh**)  
  Parameters for the plugin used to change credentials.  
  Need to be a valid JSON.  
  Need `enable_password_change` to true.

## Attribute Reference

- **id** (String)  
  Internal id of local domain in bastion.

## Import

Localdomain linked to application can be imported using an id made up of `<application_id>/<domain_name>`, e.g.

```shell
terraform import wallix-bastion_application_localdomain.app1dom xxxxxxxx/domlocal
```
