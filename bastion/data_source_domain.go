package bastion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainRead,
		Schema: map[string]*schema.Schema{
			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_real_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"admin_account": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ca_public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"password_change_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"password_change_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vault_plugin": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := resourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomain(ctx, d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get("domain_name").(string)))
	}
	cfg, err := readDomainOptions(ctx, id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDomain(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDomain(d *schema.ResourceData, jsonData jsonDomain) {
	if tfErr := d.Set("domain_real_name", jsonData.DomainRealName); tfErr != nil {
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
	if tfErr := d.Set("vault_plugin", jsonData.VaultPlugin); tfErr != nil {
		panic(tfErr)
	}
}
