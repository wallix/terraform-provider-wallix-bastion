package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceExternalAuthTacacs() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceExternalAuthTacacsRead,
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

func dataSourceExternalAuthTacacsVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_externalauth_tacacs not available with api version %s", version)
}

func dataSourceExternalAuthTacacsRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceExternalAuthTacacsVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceExternalAuthTacacs(ctx, d.Get(skAuthenticationName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get(skAuthenticationName).(string)))
	}
	cfg, err := readExternalAuthTacacsOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceExternalAuthTacacs(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceExternalAuthTacacs(d *schema.ResourceData, jsonData jsonExternalAuthTacacs) {
	if tfErr := d.Set(skAuthenticationName, jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skHost, jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPort, jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skUsePrimaryAuthDomain, jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
