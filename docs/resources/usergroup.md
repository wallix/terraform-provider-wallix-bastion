---
layout: "wallix-bastion"
page_title: "wallix-bastion: wallix-bastion_usergroup"
sidebar_current: "docs-wallix-bastion-resource-usergroup"
description: |-
  Create a usergroup
---

# wallix-bastion_usergroup

Provides a usergroup resource.

## Argument Reference

The following arguments are supported:

* `group_name` - (Required, Forces new resource)(`String`) The group name.
* `timeframes` - (Required)(`ListOfString`) The group timeframe(s).
* `description` - (Optional)(`String`) The group description.
* `profile` - (Optional)(`String`) The group profile.
* `restrictions` - (Optional)(`NestedBlock`) The group restrictions. Can be specified multiple times for each restrictions to declare.
  * `action` - (Required)(`String`) The restriction type. Need to be 'kill' or 'notify'.
  * `rules` - (Required)(`String`) The restriction rules.
  * `subprotocol` - (Required)(`String`) The restriction subprotocol.. Need to be 'SSH_SHELL_SESSION', 'SSH_REMOTE_COMMAND', 'SSH_SCP_UP', 'SSH_SCP_DOWN', 'SFTP_SESSION', 'RLOGIN', 'TELNET' or 'RDP'.
* `users` - (Optional)(`ListOfString`) The users in the group. **It's a attributes when not set**

## Attribute Reference
* `id` - (`String`) Internal id of usergroup in bastion.

## Import

Usergroup can be imported using an id made up of `<group_name>`, e.g.

```
$ terraform import wallix-bastion_user.staff staff
```
