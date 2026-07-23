package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainAccountCredentialRead,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
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
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"propagate_credential_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_domain_account_credential not available with api version %s", version)
}

func dataSourceDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomainAccountCredential(ctx,
		d.Get("domain_id").(string), d.Get("account_id").(string), d.Get("type").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s doesn't exists",
			d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string)))
	}
	cfg, err := readDomainAccountCredentialOptions(ctx,
		d.Get("domain_id").(string), d.Get("account_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDomainAccountCredential(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("public_key", jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
