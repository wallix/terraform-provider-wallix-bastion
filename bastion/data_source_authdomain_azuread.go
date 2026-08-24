package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthDomainAzureAD() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthDomainAzureADRead,
		Schema: map[string]*schema.Schema{
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAuthDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
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
			"entity_id": {
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
			skCertificate: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skIsDefault: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skPassphrase: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skSecondaryAuth: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceAuthDomainAzureADVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authdomain_azuread not available with api version %s", version)
}

func dataSourceAuthDomainAzureADRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthDomainAzureADVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainAzureAD(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get(skDomainName).(string)))
	}
	cfg, err := readAuthDomainAzureADOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthDomainAzureAD(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthDomainAzureAD(d *schema.ResourceData, jsonData jsonAuthDomainAzureAD) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAuthDomainName, jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("client_id", jsonData.ClientID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultLanguage, jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDefaultEmailDomain, jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("entity_id", jsonData.EntityID); tfErr != nil {
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
	if tfErr := d.Set(skIsDefault, jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	// private_key hidden on API
	if tfErr := d.Set(skSecondaryAuth, jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
}
