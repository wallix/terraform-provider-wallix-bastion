package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExternalAuthRadius() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExternalAuthRadiusRead,
		Schema: map[string]*schema.Schema{
			skAuthenticationName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skHost: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skTimeout: {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skUsePrimaryAuthDomain: {
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceExternalAuthRadiusVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_externalauth_radius not available with api version %s", version)
}

func dataSourceExternalAuthRadiusRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceExternalAuthRadiusVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthRadius(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get(skAuthenticationName).(string)))
	}
	cfg, err := readExternalAuthRadiusOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceExternalAuthRadius(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceExternalAuthRadius(d *schema.ResourceData, jsonData jsonExternalAuthRadius) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skTimeout, jsonData.Timeout); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
