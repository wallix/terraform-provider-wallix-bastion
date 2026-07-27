package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceAuthDomainMapping() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAuthDomainMappingRead,
		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"user_group": {
				Type:     schema.TypeString,
				Required: true,
			},
			"external_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAuthDomainMappingVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_authdomain_mapping not available with api version %s", version)
}

func dataSourceAuthDomainMappingRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceAuthDomainMappingVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceAuthDomainMapping(ctx, d.Get("domain_id").(string), d.Get("user_group").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("auth domain mapping for user_group %s on domain_id %s doesn't exists",
			d.Get("user_group").(string), d.Get("domain_id").(string)))
	}
	cfg, err := readAuthDomainMappingOptions(ctx, d.Get("domain_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceAuthDomainMapping(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceAuthDomainMapping(d *schema.ResourceData, jsonData jsonAuthDomainMapping) {
	if tfErr := d.Set("user_group", jsonData.UserGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("external_group", jsonData.ExternalGroup); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain", jsonData.Domain); tfErr != nil {
		panic(tfErr)
	}
}
