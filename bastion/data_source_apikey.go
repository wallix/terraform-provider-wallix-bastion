package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAPIKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAPIKeyRead,
		Schema: map[string]*schema.Schema{
			"apikey_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ip_limitation": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"apikey": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAPIKeyVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_apikey not available with api version %s", version)
}

func dataSourceAPIKeyRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAPIKeyVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAPIKey(ctx, d.Get("apikey_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("apikey_name %s doesn't exists", d.Get("apikey_name").(string)))
	}
	cfg, err := readAPIKeyOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAPIKey(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAPIKey(d *schema.ResourceData, jsonData jsonAPIKey) {
	if tfErr := d.Set("apikey_name", jsonData.APIKeyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_limitation", jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("apikey", jsonData.APIKey); tfErr != nil {
		panic(tfErr)
	}
}
