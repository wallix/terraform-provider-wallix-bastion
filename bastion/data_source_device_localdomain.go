package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeviceLocalDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceLocalDomainRead,
		Schema: map[string]*schema.Schema{
			"device_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"admin_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_private_key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"passphrase": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"password_change_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_change_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_change_plugin_parameters": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceDeviceLocalDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device_localdomain not available with api version %s", version)
}

func dataSourceDeviceLocalDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomain(ctx, d.Get("device_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s doesn't exists",
			d.Get("domain_name").(string), d.Get("device_id").(string)))
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, d.Get("device_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceLocalDomain(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceLocalDomain(d *schema.ResourceData, jsonData jsonDeviceLocalDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("admin_account", jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("ca_public_key", jsonData.CAPublicKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("enable_password_change", jsonData.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_policy", jsonData.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("password_change_plugin", jsonData.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
}
