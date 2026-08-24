package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDomainAccount() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceDomainAccountRead,
		Schema: map[string]*schema.Schema{
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
			"auto_change_ssh_key": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"certificate_validity": {
				Type:     schema.TypeString,
				Computed: true,
			},
			skCheckoutPolicy: {
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
						skPublicKey: {
							Type:     schema.TypeString,
							Computed: true,
						},
						skType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			skDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			skDomainPasswordChange: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"resources": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceDomainAccountVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_domain_account not available with api version %s", version)
}

func dataSourceDomainAccountRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceDomainAccountVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceDomainAccount(ctx, d.Get(skDomainID).(string), d.Get(skAccountName).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf("account_name %s on domain_id %s doesn't exists",
			d.Get(skAccountName).(string), d.Get(skDomainID).(string)))
	}
	cfg, err := readDomainAccountOptions(ctx, d.Get(skDomainID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceDomainAccount(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceDomainAccount(d *schema.ResourceData, jsonData jsonDomainAccount) {
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
				"id":        v.ID,
				skPublicKey: v.PublicKey,
				skType:      v.Type,
			}
		}
	}
	if tfErr := d.Set("credentials", credentials); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDescription, jsonData.Description); tfErr != nil {
		panic(tfErr)
	}
	if tfErr := d.Set(skDomainPasswordChange, jsonData.DomainPasswordChange); tfErr != nil {
		panic(tfErr)
	}
	if jsonData.Resources == nil {
		if tfErr := d.Set("resources", []string{}); tfErr != nil {
			panic(tfErr)
		}
	} else {
		if tfErr := d.Set("resources", *jsonData.Resources); tfErr != nil {
			panic(tfErr)
		}
	}
}
