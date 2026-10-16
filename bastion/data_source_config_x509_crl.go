package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// config_x509_crl is a singleton on the Bastion (identified by the hardcoded ID "x509CRLConfig",
// the same one used by resourceConfigX509CRL), so this data source has no lookup field.
func dataSourceConfigX509CRL() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceConfigX509CRLRead,
		Schema: map[string]*schema.Schema{
			skCRLCheck: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skCRLAutoFetch: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceConfigX509CRLVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_config_x509_crl not available with api version %s", version)
}

func dataSourceConfigX509CRLRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceConfigX509CRLVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	cfg, err := readConfigX509CRLOptions(ctx, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillConfigX509CRL(d, cfg)
	d.SetId(configX509CRLID)

	return nil
}
