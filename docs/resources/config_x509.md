# wallix-bastion_config_x509 Resource

Provides a X509 resource.

## Example Usage

```hcl
# Configure the X509 authentication and/or change GUI and API certificates
resource "wallix-bastion_config_x509" "acme-cert" {
  ca_certificate     = file("${path.root}/chain1.pem")
  server_private_key = file("${path.root}/privkey1.pem")
  server_public_key  = file("${path.root}/cert1.pem")
  enable             = true
}
```

## Argument Reference

The following arguments are supported:

- **ca_certificate** (Optional, String)  
  The ca for users authentication
- **server_private_key** (Required, String)  
  The server certificate private key
- **server_public_key** (Required, String)  
  The server certificate public key
- **enable** (Optional, Bool)  
  Whether or not enable X509 users authentication

## Attribute Reference

- **id** (String)
  Internal id of X509 config (only in Tfstate since the API does not provide any)
- **ca_certificate** (String)
  The server X509 ca certificate for users authentication
- **server_public_key** (String)
  The server x509 public certificate
- **enable** (String)
  Whether or not the X509 users authentication is enabled

## Import

X509 config can be imported using any id (in Tfstate it will always be x509Config ) e.g.

```shell
terraform import wallix-bastion_device.acme-cert myx509
```
