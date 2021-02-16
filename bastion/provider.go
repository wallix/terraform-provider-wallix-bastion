package bastion

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	versionValidate3_3 = "v3.3"
)

// Provider wallix-bastion for terraform.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"ip": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_HOST", nil),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_PORT", 443),
			},
			"token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_TOKEN", nil),
			},
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_USER", "admin"),
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_API_VERSION", "v3.3"),
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"wallix-bastion_device":                                resourceDevice(),
			"wallix-bastion_device_localdomain":                    resourceDeviceLocalDomain(),
			"wallix-bastion_device_localdomain_account":            resourceDeviceLocalDomainAccount(),
			"wallix-bastion_device_localdomain_account_credential": resourceDeviceLocalDomainAccountCredential(),
			"wallix-bastion_externalauth_ldap":                     resourceExternalAuthLdap(),
			"wallix-bastion_user":                                  resourceUser(),
			"wallix-bastion_usergroup":                             resourceUserGroup(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	config := Config{
		bastionAPIVersion: d.Get("api_version").(string),
		bastionIP:         d.Get("ip").(string),
		bastionPort:       d.Get("port").(int),
		bastionToken:      d.Get("token").(string),
		bastionUser:       d.Get("user").(string),
	}

	return config.Client()
}
