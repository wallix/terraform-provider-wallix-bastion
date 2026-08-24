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
			skDeviceID: {
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
			skCAPublicKey: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCAPrivateKey: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skEnablePasswordChange: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skPassphrase: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
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
	id, ex, err := searchResourceDeviceLocalDomain(ctx, d.Get(skDeviceID).(string), d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s on device_id %s doesn't exists",
			d.Get(skDomainName).(string), d.Get(skDeviceID).(string)))
	}
	cfg, err := readDeviceLocalDomainOptions(ctx, d.Get(skDeviceID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceLocalDomain(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceLocalDomain(d *schema.ResourceData, jsonData jsonDeviceLocalDomain) {
	if tfErr := d.Set(skDomainName, jsonData.DomainName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAdminAccount, jsonData.AdminAccount); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCAPublicKey, jsonData.CAPublicKey); tfErr != nil {
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
