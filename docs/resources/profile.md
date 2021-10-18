# wallix-bastion_profile Resource

Provides a profile resource.

## Example Usage

```hcl
resource wallix-bastion_profile example {
  profile_name = "example"
  gui_features {
    wab_audit      = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    system_audit   = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
  gui_transmission {
    system_audit   = "view"
    approval       = "view"
    authorizations = "view"
    devices        = "view"
    target_groups  = "view"
    user_groups    = "view"
    users          = "view"
    wab_settings   = "view"
  }
}
```

## Argument Reference

The following arguments are supported:

- **profile_name** (Required, String, Forces new resource)  
  The profile name.
- **gui_features** (Required, Block)  
  GUI features.
  - **wab_audit** (Optional, String)  
    Need to be `view`.
  - **system_audit** (Optional, String)  
    Need to be `view`.
  - **users** (Optional, String)  
    Need to be `view` or `modify`.
  - **user_groups** (Optional, String)  
    Need to be `view` or `modify`.
  - **devices** (Optional, String)  
    Need to be `view` or `modify`.
  - **target_groups** (Optional, String)  
    Need to be `view` or `modify`.
  - **authorizations** (Optional, String)  
    Need to be `view` or `modify`.
  - **profiles** (Optional, String)  
    Need to be `modify`.
  - **wab_settings** (Optional, String)  
    Need to be `view` or `modify`.
  - **system_settings** (Optional, String)  
    Need to be `modify`.
  - **backup** (Optional, String)  
    Need to be `execute`.
  - **approval** (Optional, String)  
    Need to be `view` or `modify`.
  - **credential_recovery** (Optional, String)  
    Need to be `execute`.
- **gui_transmission** (Required, Block)  
  GUI transmission.
  - **system_audit** (Optional, String)  
    Need to be `view`.
  - **users** (Optional, String)  
    Need to be `view` or `modify`.
  - **user_groups** (Optional, String)  
    Need to be `view` or `modify`.
  - **devices** (Optional, String)  
    Need to be `view` or `modify`.
  - **target_groups** (Optional, String)  
    Need to be `view` or `modify`.
  - **authorizations** (Optional, String)  
    Need to be `view` or `modify`.
  - **profiles** (Optional, String)  
    Need to be `modify`.
  - **wab_settings** (Optional, String)  
    Need to be `view` or `modify`.
  - **system_settings** (Optional, String)  
    Need to be `modify`.
  - **backup** (Optional, String)  
    Need to be `execute`.
  - **approval** (Optional, String)  
    Need to be `view` or `modify`.
  - **credential_recovery** (Optional, String)  
    Need to be `execute`.
- **description** (Optional, String)  
  The profile description.
- **ip_limitation** (Optional, String)  
  The profile ip limitation.  
  Format is an IPv4 address, subnet or host name.
- **target_access** (Optional, Boolean)  
  Target access.
- **target_groups_limitation** (Optional, Block)  
  Activation of target groups limitation.
  - **default_target_group** (Required, String)  
    Default target group.
  - **target_groups** (Required, List of String)  
    Target groups.
- **user_groups_limitation** (Optional, Block)  
  Activation of user groups limitation.
  - **user_groups** (Required, List of String)  
    User groups.

## Attribute Reference

- **id** (String)  
  Internal id of profile in bastion.

## Import

Profile can be imported using an id made up of `<profile_name>`, e.g.

```shell
terraform import wallix-bastion_profile.example example
```
