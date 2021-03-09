# wallix-bastion_cluster Resource

Provides a global cluster resource.

## Argument Reference

The following arguments are supported:

* `cluster_name` - (Required)(`String`) The cluster name.
* `accounts` - (Optional)(`ListOfString`) The cluster targets.
* `account_mappings` - (Optional)(`ListOfString`) The cluster targets with account mapping.
* `description` - (Optional)(`String`) The cluster description.
* `interactive_logins` - (Optional)(`ListOfString`) The cluster targets with interactive login.

## Attribute Reference

* `id` - (`String`) Internal id of cluster in bastion.

## Import

Cluster can be imported using an id made up of `<cluster_name>`, e.g.

```
$ terraform import wallix-bastion_cluster.example example
```
