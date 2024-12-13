package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthDomainAD() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthDomainADRead,
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auth_domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_email_domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_language": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"external_auths": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"language_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAuthDomainAdVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authdomain not available with api version %s", version)
}

func dataSourceAuthDomainADRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthDomainAdVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get("domain_name").(string)))
	}
	cfg, err := readAuthDomainADOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthDomainAD(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthDomainAD(d *schema.ResourceData, jsonData jsonAuthDomainAD) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_auths", jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
}
