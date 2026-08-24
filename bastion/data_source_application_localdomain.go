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
			skApplicationID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAdminAccount: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skEnablePasswordChange: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skPasswordChangePolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPasswordChangePlugin: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skPasswordChangePluginParameters: {
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
		d.Get(skApplicationID).(string), d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on application_id %s doesn't exists",
			d.Get(skDomainName).(string), d.Get(skApplicationID).(string)))
	}
	cfg, err := readApplicationLocalDomainOptions(ctx, d.Get(skApplicationID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplicationLocalDomain(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplicationLocalDomain(d *schema.ResourceData, jsonData jsonApplicationLocalDomain) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAdminAccount, jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skEnablePasswordChange, jsonData.EnablePasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPasswordChangePolicy, jsonData.PasswordChangePolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skPasswordChangePlugin, jsonData.PasswordChangePlugin); tfErr != nil {
		panic(tfErr)
	}
}
