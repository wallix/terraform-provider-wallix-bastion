package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeviceLocalDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceLocalDomainAccountCredentialRead,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
			},
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
		},
	}
}

func dataSourceDeviceLocalDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device_localdomain_account_credential "+
		"not available with api version %s", version)
}

func dataSourceDeviceLocalDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomainAccountCredential(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), d.Get("type").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("credential type %s on account_id %s, domain_id %s, device_id %s doesn't exists",
			d.Get("type").(string), d.Get("account_id").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	cfg, err := readDeviceLocalDomainAccountCredentialOptions(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceLocalDomainAccountCredential(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceLocalDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set("type", jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("public_key", jsonData.PublicKey); tfErr != nil {
		panic(tfErr)
	}
}
