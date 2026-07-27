package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDeviceLocalDomainAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDeviceLocalDomainAccountRead,
		Schema: map[string]*schema.Schema{
			"device_id": {
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
			"auto_change_ssh_key": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"certificate_validity": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"checkout_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"credentials": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"public_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_password_change": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"services": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDeviceLocalDomainAccountVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_device_localdomain_account not available with api version %s", version)
}

func dataSourceDeviceLocalDomainAccountRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDeviceLocalDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDeviceLocalDomainAccount(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), d.Get("account_name").(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s, device_id %s doesn't exists",
			d.Get("account_name").(string), d.Get("domain_id").(string), d.Get("device_id").(string)))
	}
	cfg, err := readDeviceLocalDomainAccountOptions(ctx,
		d.Get("device_id").(string), d.Get("domain_id").(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDeviceLocalDomainAccount(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDeviceLocalDomainAccount(d *schema.ResourceData, jsonData jsonDeviceLocalDomainAccount) {
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
	if tfErr := d.Set("auto_change_ssh_key", jsonData.AutoChangeSSHKey); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("certificate_validity", jsonData.CertificateValidity); tfErr != nil {
		panic(tfErr)
	}
	credentials := make([]map[string]interface{}, 0)
	if jsonData.Credentials != nil {
		credentials = make([]map[string]interface{}, len(*jsonData.Credentials))
		for i, v := range *jsonData.Credentials {
			credentials[i] = map[string]interface{}{
				"id":         v.ID,
				"public_key": v.PublicKey,
				"type":       v.Type,
			}
		}
	}
	if tfErr := d.Set("credentials", credentials); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("description", jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("domain_password_change", jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set("services", jsonData.Services); tfErr != nil {
		panic(tfErr)
	}
}
