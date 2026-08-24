package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExternalAuthLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExternalAuthLdapRead,
		Schema: map[string]*schema.Schema{
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"cn_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ldap_base": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"login_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			skTimeout: {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			skCACertificate: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCertificate: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_active_directory": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_anonymous_access": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_protected_user": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_ssl": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_starttls": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPassphrase: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skUsePrimaryAuthDomain: {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceExternalAuthLdapVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_externalauth_ldap not available with api version %s", version)
}

func dataSourceExternalAuthLdapRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceExternalAuthLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthLdap(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get(skAuthenticationName).(string)))
	}
	cfg, err := readExternalAuthLdapOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceExternalAuthLdap(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceExternalAuthLdap(d *schema.ResourceData, jsonData jsonExternalAuthLdap) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("cn_attribute", jsonData.CNAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ldap_base", jsonData.LDAPBase); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("login", jsonData.Login); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("login_attribute", jsonData.LoginAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skTimeout, jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCACertificate, jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_active_directory", jsonData.IsActiveDirectory); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_anonymous_access", jsonData.IsAnonymousAccess); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_protected_user", jsonData.IsProtectedUser); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_ssl", jsonData.IsSSL); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_starttls", jsonData.IsStartTLS); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
