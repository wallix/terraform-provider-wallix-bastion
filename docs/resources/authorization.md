# wallix-bastion_authorization Resource

Provides a authorization resource.

## Example Usage

```hcl
# Configure an authorization
resource wallix-bastion_authorization auth {
  authorization_name = "authName"
  user_group         = "userGroup"
  target_group       = "targetGroup"
  authorize_sessions = true
  subprotocols = [
    "RDP_CLIPBOARD_UP", "RDP_CLIPBOARD_DOWN", "RDP_PRINTER", "RDP_COM_PORT", "RDP_DRIVE", "RDP_SMARTCARD", "RDP_CLIPBOARD_FILE", "RDP_AUDIO_OUTPUT",
    "SSH_SHELL_SESSION", "SSH_REMOTE_COMMAND", "SSH_SCP_UP", "SSH_SCP_DOWN", "SSH_X11", "SSH_DIRECT_TCPIP", "SSH_REVERSE_TCPIP", "SSH_AUTH_AGENT",
    "SFTP_SESSION",
    "RDP",
    "VNC",
    "TELNET",
    "RLOGIN",
    "RAWTCPIP",
  ]
}
```

## Argument Reference

The following arguments are supported:

-> **Note:** At least one of `authorize_password_retrieval` or `authorize_sessions` arguments is required.

- **authorization_name** (Required, String)  
  The authorization name.
- **user_group** (Required, String, Forces new resource)  
  The user group.
- **target_group** (Required, String, Force new resource)  
  The target group.
- **description** (Optional, String)  
  The authorization description.
- **authorize_password_retrieval** (Optional, Boolean)  
  Authorize password retrieval.
- **authorize_sessions** (Optional, Boolean)  
  Authorize sessions via proxies.
  `subprotocols` need to be set.
- **subprotocols** (Optional, List of String)  
  The authorization subprotocols.  
- **is_critical** (Optional, Boolean)  
  Define if it's critical.
- **is_recorded** (Optional, Boolean)  
  Define if it's recorded.
- **approval_required** (Optional, Boolean)  
  Approval is required to connect to targets.
  `approvers` need to be set.
- **approvers** (Optional, List of String)  
  The approvers user groups.  
  `approval_required` need to be set.
- **active_quorum** (Optional, Number)  
  The quorum for active periods (-1: approval workflow with automatic approval, 0: no approval workflow (direct connection), > 0: quorum to reach).  
  Defaults to `-1`.
- **inactive_quorum** (Optional, Number)  
  The quorum for inactive periods (-1: approval workflow with automatic approval, 0: no connection allowed, > 0: quorum to reach).  
  Defaults to `-1`.
- **approval_timeout** (Optional, Number)  
  Set a timeout in minutes after which the approval will be automatically closed info connection has been initiated (i.e. the user won't be able to connect). 0: no timeout.
- **has_comment** (Optional, Boolean)  
  Comment is allowed in approval.
- **has_ticket** (Optional, Boolean)  
  Ticket is allowed in approval.
- **mandatory_comment** (Optional, Boolean)  
  Comment is mandatory in approval.
- **mandatory_ticket** (Optional, Boolean)  
  Ticket is mandatory in approval.
- **single_connection** (Optional, Boolean)  
  Limit to one single connection during the approval period (i.e. if the user disconnects, he will not be allowed to start a new session during the original requested time).

## Attribute Reference

- **id** (String)  
  Internal id of authorization in bastion.

## Import

Authorization can be imported using an id made up of `<authorization_name>`, e.g.

```shell
terraform import wallix-bastion_authorization.auth authName
```
