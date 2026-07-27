package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"golang.org/x/mod/semver"
)

// config_wsm is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "wsmConfig", the same one
// used by resourceConfigWSM). There is no natural per-instance name to look
// up, so this data source has no Required lookup field: it always reads the
// single existing Web Session Manager configuration.
func dataSourceConfigWSM() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigWSMRead,
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"jwe_public": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"jws_private": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"jws_public": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

// config/wsm doesn't exist before API v3.12.
func dataSourceConfigWSMVersionCheck(version string) error {
	if semver.Compare(version, VersionWallixAPI312) >= 0 {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_config_wsm not available with api version %s", version)
}

func dataSourceConfigWSMRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigWSMVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigWSMOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceConfigWSM(d, cfg)
	d.SetId("wsmConfig")

	return nil
}

func fillSourceConfigWSM(d *schema.ResourceData, jsonData jsonConfigWSM) {
	hostname := ""
	if jsonData.Hostname != nil {
		hostname = *jsonData.Hostname
	}
	if tfErr := d.Set("hostname", hostname); tfErr != nil {
		panic(tfErr)
	}
	jwePublic := ""
	if jsonData.JwePublic != nil {
		jwePublic = *jsonData.JwePublic
	}
	if tfErr := d.Set("jwe_public", jwePublic); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("jws_private", jsonData.JwsPrivate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("jws_public", jsonData.JwsPublic); tfErr != nil {
		panic(tfErr)
	}
}
