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
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAuthDomainName: {
				Type:     schema.TypeString,
				Required: true,
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
			"language_attribute": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skSecondaryAuth: {
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
	id, ex, err := searchResourceAuthDomainAD(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get(skDomainName).(string)))
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
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAuthDomainName, jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultLanguage, jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultEmailDomain, jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skExternalAuths, jsonData.ExternalAuths); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skSecondaryAuth, jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
}
