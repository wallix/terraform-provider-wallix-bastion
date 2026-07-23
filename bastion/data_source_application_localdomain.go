package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApplicationLocalDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationLocalDomainRead,
		Schema: map[string]*schema.Schema{
			"application_id": {
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
			"password_change_plugin_parameters": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceApplicationLocalDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_application_localdomain not available with api version %s", version)
}

func dataSourceApplicationLocalDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceApplicationLocalDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplicationLocalDomain(ctx,
		d.Get("application_id").(string), d.Get("domain_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on application_id %s doesn't exists",
			d.Get("domain_name").(string), d.Get("application_id").(string)))
	}
	cfg, err := readApplicationLocalDomainOptions(ctx, d.Get("application_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplicationLocalDomain(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplicationLocalDomain(d *schema.ResourceData, jsonData jsonApplicationLocalDomain) {
	if tfErr := d.Set("domain_name", jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("admin_account", jsonData.AdminAccount); tfErr != nil {
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
