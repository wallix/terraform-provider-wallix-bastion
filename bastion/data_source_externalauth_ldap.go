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
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cn_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host": {
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
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"ca_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"certificate": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"description": {
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
			"passphrase": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"password": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"use_primary_auth_domain": {
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
	id, ex, err := searchResourceExternalAuthLdap(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get("authentication_name").(string)))
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
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("cn_attribute", jsonData.CNAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
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
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("timeout", jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_certificate", jsonData.CACertificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
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
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
