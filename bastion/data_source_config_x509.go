package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// config_x509 is a singleton resource on the Bastion (a single, static
// configuration identified by the hardcoded ID "x509Config", the same one
// used by resourceConfigX509). There is no natural per-instance name to look
// up, so this data source has no Required lookup field: it always reads the
// single existing x509 configuration.
func dataSourceConfigX509() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigX509Read,
		Schema: map[string]*schema.Schema{
			"ca_certificate": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_private_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigX509VersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_config_x509 not available with api version %s", version)
}

func dataSourceConfigX509Read(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigX509VersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigX509Options(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceConfigX509(d, cfg)
	d.SetId("x509Config")

	return nil
}

func fillSourceConfigX509(d *schema.ResourceData, jsonData jsonConfigX509) {
	if tfErr := d.Set("ca_certificate", jsonData.CaCertificate); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("server_public_key", jsonData.ServerPublicKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("server_private_key", jsonData.ServerPrivateKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable", jsonData.Enable); tfErr != nil {
		panic(tfErr)
	}
}
