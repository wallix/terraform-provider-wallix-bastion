# wallix-bastion_cluster Resource

Provides a global cluster resource.

## Example Usage

```hcl
# Configure a cluster
resource "wallix-bastion_cluster" "example" {
  cluster_name = "example"
  interactive_logins = [
    "device1:RDP",
  ]
}
```

## Argument Reference

The following arguments are supported:

-> **Note:** At least one of `accounts`, `account_mappings` or `interactive_logins` arguments is required.

- **cluster_name**  (Required, String)  
  The cluster name.
- **accounts** (Optional, List of String)  
  The cluster targets.  
- **account_mappings** (Optional, List of String)  
  The cluster targets with account mapping.  
- **description** (Optional, String)  
  The cluster description.
- **interactive_logins** (Optional, List of String)  
  The cluster targets with interactive login.  

## Attribute Reference

- **id** (String)  
  Internal id of cluster in bastion.

## Import

Cluster can be imported using an id made up of `<cluster_name>`, e.g.

```shell
terraform import wallix-bastion_cluster.example example
```
