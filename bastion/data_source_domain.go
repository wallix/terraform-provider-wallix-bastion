package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomain() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainRead,
		Schema: map[string]*schema.Schema{
			skDomainName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_real_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skAdminAccount: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCAPublicKey: {
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
			skVaultPlugin: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceDomainVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_domain not available with api version %s", version)
}

func dataSourceDomainRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDomainVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomain(ctx, d.Get(skDomainName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("domain_name %s doesn't exists", d.Get(skDomainName).(string)))
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
	if tfErr := d.Set(skVaultPlugin, jsonData.VaultPlugin); tfErr != nil {
		panic(tfErr)
	}
}
