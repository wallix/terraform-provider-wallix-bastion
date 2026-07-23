package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/mod/semver"
)

func dataSourceAPIKeyV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAPIKeyV2Read,
		Schema: map[string]*schema.Schema{
			"apikey_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
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

// apikeys-v2 doesn't exist before API v3.12.
func dataSourceAPIKeyV2VersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_apikey_v2 not available with api version %s", version)
}

func dataSourceAPIKeyV2Read(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAPIKeyV2VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAPIKeyV2(ctx, d.Get("apikey_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("apikey_name %s doesn't exists", d.Get("apikey_name").(string)))
	}
	cfg, err := readAPIKeyV2Options(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAPIKeyV2(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAPIKeyV2(d *schema.ResourceData, jsonData jsonAPIKeyV2) {
	if tfErr := d.Set("apikey_name", jsonData.APIKeyName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("profile", jsonData.Profile); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ip_limitation", jsonData.IPLimitation); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("apikey", jsonData.APIKey); tfErr != nil {
		panic(tfErr)
	}
}
