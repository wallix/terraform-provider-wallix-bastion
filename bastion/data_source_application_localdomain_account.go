package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApplicationLocalDomainAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationLocalDomainAccountRead,
		Schema: map[string]*schema.Schema{
			skApplicationID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccountName: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccountLogin: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skAutoChangePassword: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skCheckoutPolicy: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDomainPasswordChange: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			skPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceApplicationLocalDomainAccountVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf(
		"data source wallix-bastion_application_localdomain_account not available with api version %s", version)
}

func dataSourceApplicationLocalDomainAccountRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceApplicationLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplicationLocalDomainAccount(ctx,
		d.Get(skApplicationID).(string), d.Get(skDomainID).(string), d.Get(skAccountName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, application_id %s doesn't exists",
			d.Get(skAccountName).(string), d.Get(skDomainID).(string), d.Get(skApplicationID).(string)))
	}
	cfg, err := readApplicationLocalDomainAccountOptions(ctx,
		d.Get(skApplicationID).(string), d.Get(skDomainID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplicationLocalDomainAccount(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplicationLocalDomainAccount(d *schema.ResourceData, jsonData jsonApplicationLocalDomainAccount) {
	if tfErr := d.Set(skAccountName, jsonData.AccountName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAccountLogin, jsonData.AccountLogin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skCheckoutPolicy, jsonData.CheckoutPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skAutoChangePassword, jsonData.AutoChangePassword); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDomainPasswordChange, jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
}
