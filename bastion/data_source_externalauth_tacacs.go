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
			"authentication_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"host": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"use_primary_auth_domain": {
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
	id, ex, err := searchResourceExternalAuthTacacs(ctx, d.Get("authentication_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("authentication_name %s doesn't exists", d.Get("authentication_name").(string)))
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
	if tfErr := d.Set("authentication_name", jsonData.AuthenticationName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("host", jsonData.Host); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("port", jsonData.Port); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("use_primary_auth_domain", jsonData.UsePrimaryAuthDomain); tfErr != nil {
		panic(tfErr)
	}
}
