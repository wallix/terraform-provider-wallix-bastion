package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// config_x509_ocsp is a singleton on the Bastion (identified by the hardcoded ID "x509OCSPConfig",
// the same one used by resourceConfigX509OCSP), so this data source has no lookup field.
func dataSourceConfigX509OCSP() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigX509OCSPRead,
		Schema: map[string]*schema.Schema{
			skOCSPCheck: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skDefaultOCSPResponder: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skOverrideOCSPResponder: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			// The proxy URL can hold credentials (http://user:pass@proxy:3128), which the API returns.
			skOCSPProxy: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceConfigX509OCSPVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_config_x509_ocsp not available with api version %s", version)
}

func dataSourceConfigX509OCSPRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigX509OCSPVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigX509OCSPOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigX509OCSP(d, cfg)
	d.SetId(configX509OCSPID)

	return nil
}
