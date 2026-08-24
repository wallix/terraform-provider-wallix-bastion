package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthDomainLdap() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthDomainLdapRead,
		Schema: map[string]*schema.Schema{
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAuthDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDefaultEmailDomain: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDefaultLanguage: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skExternalAuths: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"check_x509_san_email": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"display_name_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skIsDefault: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"language_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pubkey_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"san_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skSecondaryAuth: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"x509_condition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"x509_search_filter": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthDomainLdapVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authdomain_ldap not available with api version %s", version)
}

func dataSourceAuthDomainLdapRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthDomainLdapVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainLdap(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get(skDomainName).(string)))
	}
	cfg, err := readAuthDomainLdapOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthDomainLdap(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthDomainLdap(d *schema.ResourceData, jsonData jsonAuthDomainLdap) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAuthDomainName, jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skExternalAuths, jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultLanguage, jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultEmailDomain, jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("check_x509_san_email", jsonData.CheckX509SanEmail); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("group_attribute", jsonData.GroupAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("display_name_attribute", jsonData.DisplayNameAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("email_attribute", jsonData.EmailAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skIsDefault, jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("language_attribute", jsonData.LanguageAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("pubkey_attribute", jsonData.PubKeyAttribute); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("san_domain_name", jsonData.SanDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSecondaryAuth, jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("x509_condition", jsonData.X509Condition); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("x509_search_filter", jsonData.X509SearchFilter); tfErr != nil {
		panic(tfErr)
	}
}
