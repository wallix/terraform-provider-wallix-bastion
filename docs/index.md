# wallix-bastion Provider

## Argument Reference

The following arguments are supported in the `provider` block:

- **ip** (Required)  
  This is the target for bastion API connection (ip or dns name).  
  It can also be sourced from the `WALLIX_BASTION_HOST` environment variable.

- **token** (Optional)  
  This is the token to authenticate on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_TOKEN` environment variable.  

- **port** (Optional)  
  This is the tcp port for https connection on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_PORT` environment variable.  
  Defaults to `443`.

- **user** (Optional)  
  This is the username used to authenticate on bastion API.  
  It can also be sourced from the `WALLIX_BASTION_USER` environment variable.  

- **password** (Optional)
  This is the password used to authenticate against Bastion API.
  It can also be sourced from the `WALLIX_BASTION_PASSWORD`environment variable.


- **api_version** (Optional)  
  This is the version of api used to call api.  
  It can also be sourced from the `WALLIX_BASTION_API_VERSION` environment variable.  
  Defaults to `v3.3`.

- You have to specify either the API key **OR** the user/password couple. The latter is
  the recommanded authentication method. Create a dedicated account in the Bastion with the
  needed permissions according to which resources you plan to use.