package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthDomainSAML() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthDomainSAMLRead,
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auth_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
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
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_authn": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"secondary_auth": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"idp_initiated_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthDomainSAMLVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authdomain_saml not available with api version %s", version)
}

func dataSourceAuthDomainSAMLRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthDomainSAMLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainSAML(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get("domain_name").(string)))
	}
	cfg, err := readAuthDomainSAMLOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthDomainSAML(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthDomainSAML(d *schema.ResourceData, jsonData jsonAuthDomainSAML) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_auths", jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("label", jsonData.Label); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("force_authn", jsonData.ForceAuthn); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("is_default", jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_initiated_url", jsonData.IdpInitiatedURL); tfErr != nil {
		panic(tfErr)
	}
}
