# wallix-bastion_usergroup Resource

Provides a usergroup resource.

## Argument Reference

The following arguments are supported:

- **group_name** (Required, String)  
  The group name.
- **timeframes** (Required, List of String)  
  The group timeframe(s).
- **description** (Optional, String)  
  The group description.
- **profile** (Optional, String)  
  The group profile.
- **restrictions** (Optional, Set of Block)  
  The group restrictions.  
  Can be specified multiple times for each restriction to declare.
  - **action** (Required, String)  
  The restriction type.  
  Need to be `kill` or `notify`.
  - **rules** (Required, String)  
  The restriction rules.
  - **subprotocol** (Required, String)  
  The restriction subprotocol.  
  Need to be `SSH_SHELL_SESSION`, `SSH_REMOTE_COMMAND`, `SSH_SCP_UP`, `SSH_SCP_DOWN`, `SFTP_SESSION`, `RLOGIN`, `TELNET` or `RDP`.
- **users** (Optional, List of String`, **It's a attributes when not set**)  
  The users in the group.

## Attribute Reference

- **id** (String)  
  Internal id of usergroup in bastion.

## Import

Usergroup can be imported using an id made up of `<group_name>`, e.g.

```shell
terraform import wallix-bastion_usergroup.staff staff
```
