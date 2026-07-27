package bastion

import (
	"context"
	"math"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	VersionWallixAPI38  = "v3.8"
	VersionWallixAPI312 = "v3.12"
)

func defaultVersionsValid() []string {
	return []string{
		VersionWallixAPI38,
		VersionWallixAPI312,
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
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_USER", nil),
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_PORT", 443),
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_TOKEN", nil),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_PASSWORD", nil),
			},
			"api_version": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_BASTION_API_VERSION", VersionWallixAPI38),
			},
			"session_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_SESSION_TIMEOUT", 120),
				Description: "Session timeout in seconds (default: 120)",
			},
			"csrf_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_CSRF_ENABLED", true),
				Description: "Enable CSRF token protection (default: true)",
			},
			"insecure_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("WALLIX_INSECURE_SKIP_VERIFY", false),
				Description: "Skip TLS certificate verification (default: false). " +
					"Only for development with self-signed certificates.",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"wallix-bastion_apikey":                                     dataSourceAPIKey(),
			"wallix-bastion_apikey_v2":                                  dataSourceAPIKeyV2(),
			"wallix-bastion_application":                                dataSourceApplication(),
			"wallix-bastion_application_localdomain":                    dataSourceApplicationLocalDomain(),
			"wallix-bastion_application_localdomain_account":            dataSourceApplicationLocalDomainAccount(),
			"wallix-bastion_application_localdomain_account_credential": dataSourceApplicationLocalDomainAccountCredential(),
			"wallix-bastion_authdomain_ad":                              dataSourceAuthDomainAD(),
			"wallix-bastion_authdomain_azuread":                         dataSourceAuthDomainAzureAD(),
			"wallix-bastion_authdomain_ldap":                            dataSourceAuthDomainLdap(),
			"wallix-bastion_authdomain_mapping":                         dataSourceAuthDomainMapping(),
			"wallix-bastion_authdomain_saml":                            dataSourceAuthDomainSAML(),
			"wallix-bastion_authorization":                              dataSourceAuthorization(),
			"wallix-bastion_certificate_authority":                      dataSourceCertificateAuthority(),
			"wallix-bastion_checkout_policy":                            dataSourceCheckoutPolicy(),
			"wallix-bastion_cluster":                                    dataSourceCluster(),
			"wallix-bastion_configoption":                               dataSourceConfigoption(),
			"wallix-bastion_config_smtp":                                dataSourceConfigSMTP(),
			"wallix-bastion_config_wsm":                                 dataSourceConfigWSM(),
			"wallix-bastion_config_x509":                                dataSourceConfigX509(),
			"wallix-bastion_connection_message":                         dataSourceConnectionMessage(),
			"wallix-bastion_connection_policy":                          dataSourceConnectionPolicy(),
			"wallix-bastion_device":                                     dataSourceDevice(),
			"wallix-bastion_device_localdomain":                         dataSourceDeviceLocalDomain(),
			"wallix-bastion_device_localdomain_account":                 dataSourceDeviceLocalDomainAccount(),
			"wallix-bastion_device_localdomain_account_credential":      dataSourceDeviceLocalDomainAccountCredential(),
			"wallix-bastion_device_service":                             dataSourceDeviceService(),
			"wallix-bastion_domain":                                     dataSourceDomain(),
			"wallix-bastion_domain_account":                             dataSourceDomainAccount(),
			"wallix-bastion_domain_account_credential":                  dataSourceDomainAccountCredential(),
			"wallix-bastion_encryption":                                 dataSourceEncryption(),
			"wallix-bastion_externalauth_kerberos":                      dataSourceExternalAuthKerberos(),
			"wallix-bastion_externalauth_ldap":                          dataSourceExternalAuthLdap(),
			"wallix-bastion_externalauth_radius":                        dataSourceExternalAuthRadius(),
			"wallix-bastion_externalauth_saml":                          dataSourceExternalAuthSaml(),
			"wallix-bastion_externalauth_tacacs":                        dataSourceExternalAuthTacacs(),
			"wallix-bastion_local_password_policy":                      dataSourceLocalPasswordPolicy(),
			"wallix-bastion_notification":                               dataSourceNotification(),
			"wallix-bastion_passwordchangepolicy":                       dataSourcePasswordChangePolicy(),
			"wallix-bastion_profile":                                    dataSourceProfile(),
			"wallix-bastion_targetgroup":                                dataSourceTargetGroup(),
			"wallix-bastion_timeframe":                                  dataSourceTimeframe(),
			"wallix-bastion_user":                                       dataSourceUser(),
			"wallix-bastion_usergroup":                                  dataSourceUserGroup(),
			"wallix-bastion_version":                                    dataSourceVersion(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"wallix-bastion_apikey":                                     resourceAPIKey(),
			"wallix-bastion_apikey_v2":                                  resourceAPIKeyV2(),
			"wallix-bastion_application":                                resourceApplication(),
			"wallix-bastion_application_localdomain":                    resourceApplicationLocalDomain(),
			"wallix-bastion_application_localdomain_account":            resourceApplicationLocalDomainAccount(),
			"wallix-bastion_application_localdomain_account_credential": resourceApplicationLocalDomainAccountCredential(),
			"wallix-bastion_authdomain_ad":                              resourceAuthDomainAD(),
			"wallix-bastion_authdomain_azuread":                         resourceAuthDomainAzureAD(),
			"wallix-bastion_authdomain_ldap":                            resourceAuthDomainLdap(),
			"wallix-bastion_authdomain_mapping":                         resourceAuthDomainMapping(),
			"wallix-bastion_authdomain_saml":                            resourceAuthDomainSAML(),
			"wallix-bastion_authorization":                              resourceAuthorization(),
			"wallix-bastion_checkout_policy":                            resourceCheckoutPolicy(),
			"wallix-bastion_certificate_authority":                      resourceCertificateAuthority(),
			"wallix-bastion_cluster":                                    resourceCluster(),
			"wallix-bastion_config_smtp":                                resourceConfigSMTP(),
			"wallix-bastion_config_wsm":                                 resourceConfigWSM(),
			"wallix-bastion_config_x509":                                resourceConfigX509(),
			"wallix-bastion_connection_message":                         resourceConnectionMessage(),
			"wallix-bastion_connection_policy":                          resourceConnectionPolicy(),
			"wallix-bastion_device":                                     resourceDevice(),
			"wallix-bastion_device_localdomain":                         resourceDeviceLocalDomain(),
			"wallix-bastion_device_localdomain_account":                 resourceDeviceLocalDomainAccount(),
			"wallix-bastion_device_localdomain_account_credential":      resourceDeviceLocalDomainAccountCredential(),
			"wallix-bastion_device_service":                             resourceDeviceService(),
			"wallix-bastion_domain":                                     resourceDomain(),
			"wallix-bastion_domain_account":                             resourceDomainAccount(),
			"wallix-bastion_domain_account_credential":                  resourceDomainAccountCredential(),
			"wallix-bastion_externalauth_kerberos":                      resourceExternalAuthKerberos(),
			"wallix-bastion_externalauth_ldap":                          resourceExternalAuthLdap(),
			"wallix-bastion_externalauth_radius":                        resourceExternalAuthRadius(),
			"wallix-bastion_externalauth_saml":                          resourceExternalAuthSaml(),
			"wallix-bastion_externalauth_tacacs":                        resourceExternalAuthTacacs(),
			"wallix-bastion_encryption":                                 resourceEncryption(),
			"wallix-bastion_notification":                               resourceNotification(),
			"wallix-bastion_passwordchangepolicy":                       resourcePasswordChangePolicy(),
			"wallix-bastion_profile":                                    resourceProfile(),
			"wallix-bastion_targetgroup":                                resourceTargetGroup(),
			"wallix-bastion_timeframe":                                  resourceTimeframe(),
			"wallix-bastion_user":                                       resourceUser(),
			"wallix-bastion_usergroup":                                  resourceUserGroup(),
		},
		ConfigureContextFunc: configureProvider,
	}
}

func configureProvider(
	_ context.Context, d *schema.ResourceData,
) (
	interface{}, diag.Diagnostics,
) {
	config := Config{
		BastionAPIVersion:  d.Get("api_version").(string),
		BastionIP:          d.Get("ip").(string),
		BastionPort:        d.Get("port").(int),
		BastionToken:       d.Get("token").(string),
		BastionUser:        d.Get("user").(string),
		BastionPwd:         d.Get("password").(string),
		SessionTimeout:     d.Get("session_timeout").(int),
		CSRFEnabled:        d.Get("csrf_enabled").(bool),
		InsecureSkipVerify: d.Get("insecure_skip_verify").(bool),
	}

	if config.BastionIP == "" {
		return nil, diag.Errorf("missing 'ip' configuration to configure provider")
	}
	if config.BastionUser == "" {
		return nil, diag.Errorf("missing 'user' configuration to configure provider")
	}
	if config.BastionPort < 0 || config.BastionPort > math.MaxUint16 {
		return nil, diag.Errorf("invalid value %d for 'port' configuration to configure provider", config.BastionPort)
	}

	return config.Client()
}
