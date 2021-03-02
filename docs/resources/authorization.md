# wallix-bastion_authorization Resource

Provides a authorization resource.

## Argument Reference

The following arguments are supported:

* `authorization_name` - (Required)(`String`) The authorization name.
* `user_group` - (Required, ForceNew)(`String`) The user group.
* `target_group` - (Required, ForceNew)(`String`) The target group.
* `description` - (Optional)(`String`) The authorization description.
* `authorize_password_retrieval` - (Optional)(`Bool`) Authorize password retrieval.
* `authorize_sessions` - (Optional)(`Bool`) Authorize sessions via proxies. 
* `subprotocols` - (Optional)(`ListOfString`) The authorization subprotocols. Required with `authorize_sessions`.
* `is_critical` - (Optional)(`Bool`) Define if it's critical.
* `is_recorded` - (Optional)(`Bool`) Define if it's recorded.
* `approval_required` - (Optional)(`Bool`) Approval is required to connect to targets.
* `approvers` - (Optional)(`ListOfString`) The approvers user groups. Required with `approval_required`.
* `active_quorum` - (Optional)(`Int`) The quorum for active periods (-1: approval workflow with automatic approval, 0: no approval workflow (direct connection), > 0: quorum to reach). Defaults to `-1`.
* `inactive_quorum` - (Optional)(`Int`) The quorum for inactive periods (-1: approval workflow with automatic approval, 0: no connection allowed, > 0: quorum to reach). Defaults to `-1`.
* `approval_timeout` - Optional)(`Int`) Set a timeout in minutes after which the approval will be automatically closed if no connection has been initiated (i.e. the user won't be able to connect). 0: no timeout.
* `has_comment` - (Optional)(`Bool`) Comment is allowed in approval.
* `has_ticket` - (Optional)(`Bool`) Ticket is allowed in approval.
* `mandatory_comment` - (Optional)(`Bool`) Comment is mandatory in approval.
* `mandatory_ticket` - (Optional)(`Bool`) Ticket is mandatory in approval.
* `single_connection` - (Optional)(`Bool`) Limit to one single connection during the approval period (i.e. if the user disconnects, he will not be allowed to start a new session during the original requested time).

## Attribute Reference

* `id` - (`String`) Internal id of authorization in bastion.

## Import

Authorization can be imported using an id made up of `<authorization_name>`, e.g.

```
$ terraform import wallix-bastion_authorization.auth authName
```
