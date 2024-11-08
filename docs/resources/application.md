# wallix-bastion_application Resource

Provides a application resource.

## Example Usage

```hcl
# Configure an application
resource "wallix-bastion_application" "app1" {
  application_name  = "app1"
  connection_policy = "RDP"
  paths {
    target      = "Interactive@device:SSH"
    program     = "application_path"
    working_dir = "directory"
  }
  target = "cluster"
}
```

## Argument Reference

The following arguments are supported:

- **application_name** (Required, String)  
  The application name.
- **connection_policy**  (Required, String)  
  The connection policy name.
- **category** (Optional, String)  
  The application category.  
  Default to `standard`.  
  Need to be `standard` or `jumphost`.
- **application_url** (Optional, String)  
  The application url.  
  `category` need to be `jumphost`.
- **browser** (Optional, String)  
  The application browser.  
  `category` need to be `jumphost`.
- **browser_version** (Optional, String)  
  The browser version.  
  `category` need to be `jumphost`.
- **description** - (Optional, String)  
  The application description.
- **global_domains** (Optional, List of String)  
  The global domains names.  
  `category` need to be `standard`.
- **parameters** (Optional, String)  
  The application parameters.
- **paths** (Optional, Set of Block)  
  Need to be specified when `category` = `standard`,
  multiple times for each target in cluster or once if target is a device's session.
  - **target** (Required, String)  
    The application target.
  - **program** (Required, String)  
    The application path.
  - **working_dir** (Required, String)  
    The application working directory.
- **target** (Optional, String)  
  The application target/cluster name.  
  Need to be specified when `category` = `standard`

## Attribute Reference

- **id** (String)  
  Internal id of application in bastion.
- **local_domains** (List of Block)  
  List of localdomain.
  - **id** (String)  
    Internal id of local domain in bastion.
  - **domain_name** (String)  
    The domain name.
  - **description** (String)  
    The domain description.
  - **enable_password_change** (Boolean)  
    Enable the change of password on this domain.
  - **password_change_policy** (String)  
    The name of password change policy for this domain.
  - **password_change_plugin** (String)  
    The name of plugin used to change passwords on this domain.
  - **password_change_plugin_parameters** (String)  
    Parameters for the plugin used to change credentials.

## Import

Application can be imported using an id made up of `<application_name>`, e.g.

```shell
terraform import wallix-bastion_application.app1 app1
```
