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
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"force_authn": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skIsDefault: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skSecondaryAuth: {
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
	id, ex, err := searchResourceAuthDomainSAML(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get(skDomainName).(string)))
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
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAuthDomainName, jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultEmailDomain, jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultLanguage, jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skExternalAuths, jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("label", jsonData.Label); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("force_authn", jsonData.ForceAuthn); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skIsDefault, jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSecondaryAuth, jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("idp_initiated_url", jsonData.IdpInitiatedURL); tfErr != nil {
		panic(tfErr)
	}
}
