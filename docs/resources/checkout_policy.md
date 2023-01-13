# wallix-bastion_checkout_policy Resource

Provides a checkout_policy resource.

## Example Usage

```hcl
# Configure a checkout policy
resource "wallix-bastion_checkout_policy" "pol" {
  checkout_policy_name = "example"
}
```

## Argument Reference

The following arguments are supported:

- **checkout_policy_name** (Required, String)  
  The checkout policy name.
- **description** (Optional, String)  
  The checkout policy description.
- **enable_lock** (Optional, Boolean)  
  Lock on checkout.
  `duration` and `max_duration` need to be set.
- **change_credentials_at_checkin** (Optional, Boolean)  
  Change credentials at check-in.  
  `enable_lock` need to be set.
- **duration** (Optional, Number)  
  The checkout duration (in seconds).  
  `enable_lock` need to be set.
- **extension** (Optional, Number)  
  The extension duration (in seconds).  
  `enable_lock` need to be set.
- **max_duration** (Optional, Number)  
  The max duration (in seconds).  
  `enable_lock` need to be set.

## Attribute Reference

- **id** (String)  
  Internal id of checkout policy in bastion.

## Import

Checkout policy can be imported using an id made up of `<checkout_policy_name>`, e.g.

```shell
terraform import wallix-bastion_checkout_policy.pol example
```
