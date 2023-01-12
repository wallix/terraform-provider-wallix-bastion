package bastion

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	VersionWallixAPI33 = "v3.3"
	VersionWallixAPI36 = "v3.6"
	VersionWallixAPI38 = "v3.8"
)

func defaultVersionsValid() []string {
	return []string{
		VersionWallixAPI33,
		VersionWallixAPI36,
		VersionWallixAPI38,
	}
}

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
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_API_VERSION", VersionWallixAPI33),
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"wallix-bastion_domain":  dataSourceDomain(),
			"wallix-bastion_version": dataSourceVersion(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"wallix-bastion_application":                           resourceApplication(),
			"wallix-bastion_application_localdomain":               resourceApplicationLocalDomain(),
			"wallix-bastion_application_localdomain_account":       resourceApplicationLocalDomainAccount(),
			"wallix-bastion_authdomain_ldap":                       resourceAuthDomainLdap(),
			"wallix-bastion_authdomain_mapping":                    resourceAuthDomainMapping(),
			"wallix-bastion_authorization":                         resourceAuthorization(),
			"wallix-bastion_checkout_policy":                       resourceCheckoutPolicy(),
			"wallix-bastion_cluster":                               resourceCluster(),
			"wallix-bastion_connection_policy":                     resourceConnectionPolicy(),
			"wallix-bastion_device":                                resourceDevice(),
			"wallix-bastion_device_localdomain":                    resourceDeviceLocalDomain(),
			"wallix-bastion_device_localdomain_account":            resourceDeviceLocalDomainAccount(),
			"wallix-bastion_device_localdomain_account_credential": resourceDeviceLocalDomainAccountCredential(),
			"wallix-bastion_device_service":                        resourceDeviceService(),
			"wallix-bastion_domain":                                resourceDomain(),
			"wallix-bastion_domain_account":                        resourceDomainAccount(),
			"wallix-bastion_domain_account_credential":             resourceDomainAccountCredential(),
			"wallix-bastion_externalauth_kerberos":                 resourceExternalAuthKerberos(),
			"wallix-bastion_externalauth_ldap":                     resourceExternalAuthLdap(),
			"wallix-bastion_externalauth_radius":                   resourceExternalAuthRadius(),
			"wallix-bastion_externalauth_saml":                     resourceExternalAuthSaml(),
			"wallix-bastion_externalauth_tacacs":                   resourceExternalAuthTacacs(),
			"wallix-bastion_ldapdomain":                            resourceLdapDomain(),
			"wallix-bastion_ldapmapping":                           resourceLdapMapping(),
			"wallix-bastion_profile":                               resourceProfile(),
			"wallix-bastion_targetgroup":                           resourceTargetGroup(),
			"wallix-bastion_timeframe":                             resourceTimeframe(),
			"wallix-bastion_user":                                  resourceUser(),
			"wallix-bastion_usergroup":                             resourceUserGroup(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(
	ctx context.Context, d *schema.ResourceData,
) (
	interface{}, diag.Diagnostics,
) {
	config := Config{
		bastionAPIVersion: d.Get("api_version").(string),
		bastionIP:         d.Get("ip").(string),
		bastionPort:       d.Get("port").(int),
		bastionToken:      d.Get("token").(string),
		bastionUser:       d.Get("user").(string),
	}

	return config.Client()
}
