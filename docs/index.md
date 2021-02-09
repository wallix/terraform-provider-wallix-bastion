# wallix-bastion Provider

## Argument Reference

The following arguments are supported in the `provider` block:

* `ip` - (Required) This is the target for bastion API connection (ip or dns name).  
  It can also be sourced from the `WALLIX_BASTION_HOST` environment variable.

* `token` - (Required) This is the token to authenticate on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_TOKEN` environment variable.  
  Defaults is empty.

* `port` - (Optional) This is the tcp port for https connection on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_PORT` environment variable.  
  Defaults to `443`.

* `user` - (Optional) This is the username used to authenticate on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_USER` environment variable.  
  Defaults to `admin`.

* `api_version` - (Optional) This is the version of api used to call api.  
  It can also be sourced from the `WALLIX_BASTION_API_VERSION` environment variable.  
  Defaults to `v3.3`.