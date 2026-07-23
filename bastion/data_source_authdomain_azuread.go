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
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"auth_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_id": {
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
			"entity_id": {
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
			"certificate": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_default": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"passphrase": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"secondary_auth": {
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
	id, ex, err := searchResourceAuthDomainAzureAD(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get("domain_name").(string)))
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
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auth_domain_name", jsonData.AuthDomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("client_id", jsonData.ClientID); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_language", jsonData.DefaultLanguage); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("default_email_domain", jsonData.DefaultEmailDomain); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("entity_id", jsonData.EntityID); tfErr != nil {
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
	if tfErr := d.Set("is_default", jsonData.IsDefault); tfErr != nil {
		panic(tfErr)
	}
	// private_key hidden on API
	if tfErr := d.Set("secondary_auth", jsonData.SecondaryAuth); tfErr != nil {
		panic(tfErr)
	}
}
