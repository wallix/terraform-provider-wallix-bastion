package bastion

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceApplicationLocalDomainAccountCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceApplicationLocalDomainAccountCredentialRead,
		Schema: map[string]*schema.Schema{
			skApplicationID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skDomainID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skAccountID: {
				Type:     schema.TypeString,
				Required: true,
			},
			skType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{skPassword}, false),
			},
			skPassword: {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func dataSourceApplicationLocalDomainAccountCredentialVersionCheck(version string) error {
	if slices.Contains(defaultVersionsValid(), version) {
		return nil
	}

	return fmt.Errorf("data source wallix-bastion_application_localdomain_account_credential "+
		"not available with api version %s", version)
}

func dataSourceApplicationLocalDomainAccountCredentialRead(
	ctx context.Context, d *schema.ResourceData, m interface{},
) diag.Diagnostics {
	c := m.(*Client)
	if err := dataSourceApplicationLocalDomainAccountCredentialVersionCheck(c.bastionAPIVersion); err != nil {
		return diag.FromErr(err)
	}
	id, ex, err := searchResourceApplicationLocalDomainAccountCredential(ctx,
		d.Get(skApplicationID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string),
		d.Get(skType).(string), m)
	if err != nil {
		return diag.FromErr(err)
	}
	if !ex {
		return diag.FromErr(fmt.Errorf(
			"credential type %s on account_id %s, domain_id %s, application_id %s doesn't exists",
			d.Get(skType).(string), d.Get(skAccountID).(string), d.Get(skDomainID).(string), d.Get(skApplicationID).(string)))
	}
	cfg, err := readApplicationLocalDomainAccountCredentialOptions(ctx,
		d.Get(skApplicationID).(string), d.Get(skDomainID).(string), d.Get(skAccountID).(string), id, m)
	if err != nil {
		return diag.FromErr(err)
	}
	fillSourceApplicationLocalDomainAccountCredential(d, cfg)
	d.SetId(id)

	return nil
}

func fillSourceApplicationLocalDomainAccountCredential(d *schema.ResourceData, jsonData jsonCredential) {
	if tfErr := d.Set(skType, jsonData.Type); tfErr != nil {
		panic(tfErr)
	}
}
