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
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"account_login": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_change_password": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"checkout_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"password": {
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
		d.Get("application_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, application_id %s doesn't exists",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("application_id").(string)))
	}
	cfg, err := readApplicationLocalDomainAccountOptions(ctx,
		d.Get("application_id").(string), d.Get("domain_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplicationLocalDomainAccount(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplicationLocalDomainAccount(d *schema.ResourceData, jsonData jsonApplicationLocalDomainAccount) {
	if tfErr := d.Set("account_name", jsonData.AccountName); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("account_login", jsonData.AccountLogin); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("checkout_policy", jsonData.CheckoutPolicy); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("auto_change_password", jsonData.AutoChangePassword); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_password_change", jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
}
