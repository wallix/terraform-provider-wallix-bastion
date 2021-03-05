# wallix-bastion_profile Resource

Provides a profile resource.

## Argument Reference

The following arguments are supported:

* `profile_name` - (Required, Forces new resource)(`String`) The profile name.
* `gui_features` - (Required)(`NestedBlock`) Can be specified one.
  * `wab_audit` - (Optional)(`String`) Need to be 'view'.
  * `system_audit` - (Optional)(`String`) Need to be 'view'.
  * `users` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `user_groups` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `devices` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `target_groups` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `authorizations` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `profiles` - (Optional)(`String`) Need to be 'modify'.
  * `wab_settings` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `system_settings` - (Optional)(`String`) Need to be 'modify'.
  * `backup` - (Optional)(`String`) Need to be 'execute'.
  * `approval` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `credential_recovery` - (Optional)(`String`) Need to be 'execute'.
* `gui_transmission` - (Required)(`NestedBlock`) Can be specified one.
  * `system_audit` - (Optional)(`String`) Need to be 'view'.
  * `users` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `user_groups` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `devices` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `target_groups` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `authorizations` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `profiles` - (Optional)(`String`) Need to be 'modify'.
  * `wab_settings` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `system_settings` - (Optional)(`String`) Need to be 'modify'.
  * `backup` - (Optional)(`String`) Need to be 'execute'.
  * `approval` - (Optional)(`String`) Need to be 'view' or 'modify'.
  * `credential_recovery` - (Optional)(`String`) Need to be 'execute'.
* `description` - (Optional)(`String`) The profile description.
* `ip_limitation` - (Optional)(`String`) The profile ip limitation. Format is an IPv4 address, subnet or host name.
* `target_access` - (Optional)(`Bool`) Target access.
* `target_groups_limitation` - (Optional)(`NestedBlock`) Activation of target groups limitation. Can be specified one.
  * `target_groups` - (Required)(`ListOfString`) Target groups.
  * `default_target_group` - (Optional)(`String`) Default target group.
* `user_groups_limitation` - (Optional)(`NestedBlock`) Activation of user groups limitation. Can be specified one.
  * `user_groups` - (Required)(`ListOfString`) User groups.

## Attribute Reference

* `id` - (`String`) Internal id of profile in bastion.

## Import

Profile can be imported using an id made up of `<profile_name>`, e.g.

```
$ terraform import wallix-bastion_profile.example example
```
