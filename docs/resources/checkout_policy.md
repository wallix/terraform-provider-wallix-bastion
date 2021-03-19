# wallix-bastion_checkout_policy Resource

Provides a checkout_policy resource.

## Argument Reference

The following arguments are supported:

* `checkout_policy_name` - (Required)(`String`) The checkout policy name.
* `description` - (Optional)(`String`) The checkout policy description.
* `enable_lock` - (Optional)(`Bool`) Lock on checkout.
* `change_credentials_at_checkin` - (Optional)(`Bool`) Change credentials at check-in. `enable_lock` need to be set.
* `duration` - (Optional)(`Int`) The checkout duration (in seconds). Required with `enable_lock`.
* `extension` - (Optional)(`Int`) The extension duration (in seconds). `enable_lock` need to be set.
* `max_duration` - (Optional)(`Int`) The max duration (in seconds). Required with `enable_lock`.

## Attribute Reference

* `id` - (`String`) Internal id of checkout policy in bastion.

## Import

Checkout policy can be imported using an id made up of `<checkout_policy_name>`, e.g.

```
$ terraform import wallix-bastion_checkout_policy.pol example
```
