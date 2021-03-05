# wallix-bastion_targetgroup Resource

Provides a targetgroup resource.

## Argument Reference

The following arguments are supported:

* `group_name` - (Required)(`String`) The target group name.
* `description` - (Optional)(`String`) The target group description.
* `password_retrieval_accounts` - (Optional)(`NestedBlock`) The accounts (for checkout/checkin). The accounts must exist in the Bastion. Can be specified multiple times for each account/domain to declare.
  * `account` - (Required)(`String`) The account name.
  * `domain` - (Required)(`String`) The domain name.
  * `domain_type` - (Required)(`String`) The domain type. Neet to be 'local' or 'global'.
  * `device` - (Optional)(`String`) The device name (null for an application or a global domain).
  * `application` - (Optional)(`String`) The application name (null for a device or a global domain).
* `restrictions` - (Optional)(`NestedBlock`) The group restrictions. Can be specified multiple times for each restriction to declare.
  * `action` - (Required)(`String`) The restriction type. Need to be 'kill' or 'notify'.
  * `rules` - (Required)(`String`) The restriction rules.
  * `subprotocol` - (Required)(`String`) The restriction subprotocol.. Need to be 'SSH_SHELL_SESSION', 'SSH_REMOTE_COMMAND', 'SSH_SCP_UP', 'SSH_SCP_DOWN', 'SFTP_SESSION', 'RLOGIN', 'TELNET' or 'RDP'.
* `session_accounts` - (Optional)(`NestedBlock`) The devices and applications accounts. The accounts must exist in the Bastion. Can be specified multiple times for each account/domain to declare.
  * `account` - (Required)(`String`) The account name.
  * `domain` - (Required)(`String`) The domain name.
  * `domain_type` - (Required)(`String`) The domain type: local or global.
  * `device` - (Optional)(`String`) The device name (null for an application or a global domain).
  * `service` - (Optional)(`String`) The service name (null for an application or a global domain).
  * `application` - (Optional)(`String`) The application name (null for a device or a global domain).
* `session_account_mappings` - (Optional)(`NestedBlock`) The devices/applications accounts mappings. The accounts must exist in the Bastion. Can be specified multiple times for each mapping to declare.
  * `device` - (Optional)(`String`) The device name (null for an application).
  * `service` - (Optional)(`String`) The service name (null for an application).
  * `application` - (Optional)(`String`) The application name (null for a device).
* `session_interactive_logins` - (Optional)(`NestedBlock`) The accounts on devices/applications with interactive logins. The accounts must exist in the Bastion. Can be specified multiple times for each device/application to declare.
  * `device` - (Optional)(`String`) The device name (null for an application).
  * `service` - (Optional)(`String`) The service name (null for an application).
  * `application` - (Optional)(`String`) The application name (null for a device).
* `session_scenario_accounts` - (Optional)(`NestedBlock`) The devices and applications accounts to use for scenario. The accounts must exist in the Bastion. Can be specified multiple times for each account/domain to declare.
  * `account` - (Required)(`String`) The account name.
  * `domain` - (Required)(`String`) The domain name.
  * `domain_type` - (Required)(`String`) The domain type: local or global.
  * `device` - (Optional)(`String`) The device name (null for an application or a global domain).
  * `application` - (Optional)(`String`) The application name (null for a device or a global domain).

## Attribute Reference

* `id` - (`String`) Internal id of targetgroup in bastion.

## Import

Targetgroup can be imported using an id made up of `<group_name>`, e.g.

```
$ terraform import wallix-bastion_targetgroup.group groupName
```
